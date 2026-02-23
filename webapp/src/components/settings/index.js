import React, {PureComponent} from 'react';

const PLUGIN_ID = 'com.github.reaction-notification';
const API_BASE = `/plugins/${PLUGIN_ID}/api/v1`;

export default class UserSettingsPanel extends PureComponent {
    constructor(props) {
        super(props);
        this.state = {
            enabled: true,
            excludedEmoji: [],
            newEmoji: '',
            loading: true,
            saving: false,
            error: null,
            saved: false,
        };
    }

    async componentDidMount() {
        try {
            const resp = await fetch(`${API_BASE}/prefs`, {
                method: 'GET',
                headers: {'X-Requested-With': 'XMLHttpRequest'},
            });
            if (!resp.ok) {
                throw new Error(`HTTP ${resp.status}`);
            }
            const prefs = await resp.json();
            this.setState({
                enabled: prefs.enabled,
                excludedEmoji: prefs.excluded_emoji || [],
                loading: false,
            });
        } catch (err) {
            this.setState({error: err.message, loading: false});
        }
    }

    handleToggleEnabled = () => {
        this.setState((prev) => ({enabled: !prev.enabled, saved: false}));
    };

    handleNewEmojiChange = (e) => {
        this.setState({newEmoji: e.target.value, saved: false});
    };

    handleAddEmoji = () => {
        // Strip surrounding colons if the user typed ":thumbsup:" instead of "thumbsup".
        const emoji = this.state.newEmoji.trim().replace(/^:|:$/g, '');
        if (!emoji) {
            return;
        }
        if (this.state.excludedEmoji.includes(emoji)) {
            this.setState({newEmoji: ''});
            return;
        }
        this.setState((prev) => ({
            excludedEmoji: [...prev.excludedEmoji, emoji],
            newEmoji: '',
            saved: false,
        }));
    };

    handleRemoveEmoji = (emoji) => {
        this.setState((prev) => ({
            excludedEmoji: prev.excludedEmoji.filter((e) => e !== emoji),
            saved: false,
        }));
    };

    handleKeyDown = (e) => {
        if (e.key === 'Enter') {
            this.handleAddEmoji();
        }
    };

    handleSave = async () => {
        this.setState({saving: true, error: null, saved: false});
        try {
            const resp = await fetch(`${API_BASE}/prefs`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Requested-With': 'XMLHttpRequest',
                },
                body: JSON.stringify({
                    enabled: this.state.enabled,
                    excluded_emoji: this.state.excludedEmoji,
                }),
            });
            if (!resp.ok) {
                throw new Error(`HTTP ${resp.status}`);
            }
            this.setState({saving: false, saved: true});
        } catch (err) {
            this.setState({saving: false, error: err.message});
        }
    };

    render() {
        const {loading, saving, error, saved, enabled, excludedEmoji, newEmoji} = this.state;

        if (loading) {
            return (
                <div style={{padding: '16px'}}>
                    {'Loading...'}
                </div>
            );
        }

        return (
            <div style={{padding: '16px', maxWidth: '480px'}}>
                <h4 style={{marginTop: 0}}>{'Thread Reaction Notifications'}</h4>
                <p style={{color: 'var(--center-channel-color-56)', marginBottom: '16px'}}>
                    {'Get notified when someone reacts to a message in a thread you participated in.'}
                </p>

                {/* Global enable/disable toggle */}
                <div style={{marginBottom: '16px'}}>
                    <label style={{display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer'}}>
                        <input
                            type='checkbox'
                            checked={enabled}
                            onChange={this.handleToggleEnabled}
                        />
                        {'Enable reaction notifications'}
                    </label>
                </div>

                {/* Excluded emoji list — only shown when notifications are enabled */}
                {enabled && (
                    <div style={{marginBottom: '16px'}}>
                        <label style={{display: 'block', fontWeight: 600, marginBottom: '4px'}}>
                            {'Excluded Emojis'}
                        </label>
                        <p style={{color: 'var(--center-channel-color-56)', fontSize: '12px', margin: '0 0 8px'}}>
                            {"You won't be notified for reactions using these emojis."}
                        </p>

                        {/* Tags for currently excluded emojis */}
                        {excludedEmoji.length > 0 && (
                            <div style={{display: 'flex', flexWrap: 'wrap', gap: '6px', marginBottom: '8px'}}>
                                {excludedEmoji.map((emoji) => (
                                    <span
                                        key={emoji}
                                        style={{
                                            display: 'inline-flex',
                                            alignItems: 'center',
                                            gap: '4px',
                                            padding: '2px 8px',
                                            borderRadius: '12px',
                                            background: 'var(--center-channel-color-08)',
                                            fontSize: '13px',
                                        }}
                                    >
                                        {`:`}{emoji}{`:`}
                                        <button
                                            type='button'
                                            onClick={() => this.handleRemoveEmoji(emoji)}
                                            aria-label={`Remove ${emoji}`}
                                            style={{
                                                background: 'none',
                                                border: 'none',
                                                cursor: 'pointer',
                                                padding: '0',
                                                lineHeight: 1,
                                                color: 'inherit',
                                            }}
                                        >
                                            {'×'}
                                        </button>
                                    </span>
                                ))}
                            </div>
                        )}

                        {/* Add new emoji input */}
                        <div style={{display: 'flex', gap: '8px'}}>
                            <input
                                type='text'
                                className='form-control'
                                placeholder='e.g. thumbsup or :thumbsup:'
                                value={newEmoji}
                                onChange={this.handleNewEmojiChange}
                                onKeyDown={this.handleKeyDown}
                                style={{flex: 1}}
                            />
                            <button
                                type='button'
                                className='btn btn-default'
                                onClick={this.handleAddEmoji}
                            >
                                {'Add'}
                            </button>
                        </div>
                    </div>
                )}

                {/* Save button and feedback */}
                <div style={{display: 'flex', alignItems: 'center', gap: '12px'}}>
                    <button
                        type='button'
                        className='btn btn-primary'
                        onClick={this.handleSave}
                        disabled={saving}
                    >
                        {saving ? 'Saving...' : 'Save'}
                    </button>
                    {saved && (
                        <span style={{color: 'var(--online-indicator)'}}>
                            {'Settings saved.'}
                        </span>
                    )}
                    {error && (
                        <span style={{color: 'var(--error-text)'}}>
                            {`Error: ${error}`}
                        </span>
                    )}
                </div>
            </div>
        );
    }
}
