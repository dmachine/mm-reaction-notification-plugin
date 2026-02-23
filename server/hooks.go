package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

// ReactionHasBeenAdded is called by Mattermost after a reaction is committed
// to the database. It notifies all other thread participants via DM.
func (p *Plugin) ReactionHasBeenAdded(c *plugin.Context, reaction *model.Reaction) {
	// Guard: skip reactions added by bots to prevent notification spam.
	reactor, appErr := p.API.GetUser(reaction.UserId)
	if appErr != nil {
		p.API.LogError("ReactionHasBeenAdded: GetUser (reactor) failed",
			"user_id", reaction.UserId, "err", appErr.Error())
		return
	}
	if reactor.IsBot {
		return
	}

	// Get the post that was reacted to.
	post, appErr := p.API.GetPost(reaction.PostId)
	if appErr != nil {
		p.API.LogError("ReactionHasBeenAdded: GetPost failed",
			"post_id", reaction.PostId, "err", appErr.Error())
		return
	}

	// Guard: skip reactions in DM or Group Message channels to avoid
	// sending a confusing DM-within-DM notification.
	channel, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil {
		p.API.LogError("ReactionHasBeenAdded: GetChannel failed",
			"channel_id", post.ChannelId, "err", appErr.Error())
		return
	}
	if channel.Type == model.ChannelTypeDirect || channel.Type == model.ChannelTypeGroup {
		return
	}

	// Determine the thread root ID.
	// If the reacted post is a reply, RootId points to the thread root.
	// If the reacted post is itself the root (or a standalone post), use its
	// own Id — GetPostThread will return all replies from that root.
	rootID := post.RootId
	if rootID == "" {
		rootID = post.Id
	}

	// Get all posts in the thread.
	thread, appErr := p.API.GetPostThread(rootID)
	if appErr != nil {
		p.API.LogError("ReactionHasBeenAdded: GetPostThread failed",
			"root_id", rootID, "err", appErr.Error())
		return
	}

	// Guard: if the thread has only one post (the root itself with no replies),
	// there are no other participants to notify.
	if len(thread.Posts) <= 1 {
		return
	}

	// Distributed lock: in Mattermost HA clusters, the hook may fire on
	// multiple nodes for the same event. Use KVCompareAndSet to ensure only
	// one node processes each unique reaction.
	lockKey := fmt.Sprintf("reaction_lock_%s_%s_%s",
		reaction.PostId, reaction.UserId, reaction.EmojiName)
	set, appErr := p.API.KVCompareAndSet(lockKey, nil, []byte("1"))
	if appErr != nil {
		p.API.LogError("ReactionHasBeenAdded: KVCompareAndSet failed",
			"lock_key", lockKey, "err", appErr.Error())
		// Proceed without the lock rather than silently drop notifications.
	} else if !set {
		// Another node already claimed the lock for this reaction event.
		return
	}
	// Release the lock after a brief window — we don't keep it long because
	// the reaction event should only fire once per reaction addition.
	defer p.API.KVDelete(lockKey) //nolint:errcheck

	// Collect all unique human participants in the thread, excluding the reactor.
	participantIDs := p.collectParticipants(thread, reaction.UserId)
	if len(participantIDs) == 0 {
		return
	}

	// Build the deep-link permalink to the reacted post.
	// Mattermost requires the team name in the URL: /{teamName}/pl/{postId}.
	// The teamless /pl/{postId} route resolves to "team not found" in practice.
	siteURL := ""
	if cfg := p.API.GetConfig(); cfg != nil && cfg.ServiceSettings.SiteURL != nil {
		siteURL = *cfg.ServiceSettings.SiteURL
	}
	teamName := ""
	if team, appErr := p.API.GetTeam(channel.TeamId); appErr == nil {
		teamName = team.Name
	}
	permalink := buildPermalink(siteURL, teamName, reaction.PostId)

	notificationMsg := fmt.Sprintf(
		"**@%s** reacted with :%s: to a message in a thread you participated in.\n[View thread →](%s)",
		reactor.Username,
		reaction.EmojiName,
		permalink,
	)

	// Send an ephemeral notification to each participant who has not opted out.
	for _, recipientID := range participantIDs {
		if err := p.sendNotification(recipientID, reaction.UserId, post.ChannelId, reaction.EmojiName, notificationMsg); err != nil {
			p.API.LogError("ReactionHasBeenAdded: sendNotification failed",
				"recipient_id", recipientID, "err", err.Error())
			// Continue — one failure must not block notifications to others.
		}
	}
}

// collectParticipants returns the unique set of human user IDs who have posted
// in the thread, excluding the reactor and any bot accounts.
func (p *Plugin) collectParticipants(thread *model.PostList, reactorUserID string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, post := range thread.Posts {
		uid := post.UserId
		if uid == reactorUserID {
			continue
		}
		if seen[uid] {
			continue
		}
		user, appErr := p.API.GetUser(uid)
		if appErr != nil {
			// Cannot verify the user; skip rather than notify an unknown account.
			continue
		}
		if user.IsBot {
			continue
		}
		seen[uid] = true
		result = append(result, uid)
	}
	return result
}

// sendNotification checks the recipient's preferences and, if they have not
// opted out or excluded the given emoji, sends an ephemeral post visible only
// to the recipient in the channel where the thread lives.
func (p *Plugin) sendNotification(recipientID, reactorUserID, channelID, emojiName, message string) error {
	prefs, err := p.GetUserPreferences(recipientID)
	if err != nil {
		return fmt.Errorf("GetUserPreferences: %w", err)
	}
	if !prefs.Enabled {
		return nil
	}
	for _, excluded := range prefs.ExcludedEmoji {
		if strings.EqualFold(excluded, emojiName) {
			return nil
		}
	}

	// SendEphemeralPost delivers the post only to recipientID; it is never
	// stored in the database and disappears on page refresh.
	p.API.SendEphemeralPost(recipientID, &model.Post{
		ChannelId: channelID,
		UserId:    reactorUserID,
		Message:   message,
		Type:      model.PostTypeDefault,
	})
	return nil
}

// buildPermalink constructs a deep-link URL to a specific post.
// Mattermost's /{teamName}/pl/{postId} route redirects to the correct channel
// view. The teamless /pl/{postId} variant fails with "team not found".
func buildPermalink(siteURL, teamName, postID string) string {
	if siteURL == "" || postID == "" {
		return ""
	}
	base := strings.TrimRight(siteURL, "/")
	if teamName == "" {
		return fmt.Sprintf("%s/pl/%s", base, postID)
	}
	return fmt.Sprintf("%s/%s/pl/%s", base, teamName, postID)
}
