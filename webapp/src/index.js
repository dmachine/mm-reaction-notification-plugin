import UserSettingsPanel from './components/settings';

const PLUGIN_ID = 'com.github.reaction-notification';

class Plugin {
    /**
     * initialize is called by the Mattermost webapp when the plugin is loaded.
     *
     * @param {PluginRegistry} registry - Registry for UI component hooks.
     * @param {Store}          store    - Redux store (not used here but available).
     */
    initialize(registry /* , store */) {
        // Register per-user notification preferences in the Account Settings modal.
        // Uses the PluginRegistry.registerUserSettings API (introduced in MM v7.6).
        // Users find this under: Account Settings > Notifications > Reaction Notifications
        if (registry.registerUserSettings) {
            registry.registerUserSettings({
                id: PLUGIN_ID,
                uiName: 'Reaction Notifications',
                sections: [
                    {
                        title: 'Thread Reaction Notifications',
                        component: UserSettingsPanel,
                    },
                ],
            });
        }
    }
}

// Register the plugin class with the Mattermost webapp runtime.
window.registerPlugin(PLUGIN_ID, new Plugin());
