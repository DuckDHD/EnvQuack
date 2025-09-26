package quack

// GetHappyDuck returns ASCII art for when everything is fine
func GetHappyDuck() string {
	return `   __
<(o )___   All good!
 ( ._> /
  '---'`
}

// GetAngryDuck returns ASCII art for when there are issues
func GetAngryDuck() string {
	return `   __
<(X )___   QUACK!
 ( ._> /
  '---'`
}

// GetBanner returns the main EnvQuack banner
func GetBanner() string {
	return `
 ___            ___                 _    
| __|_ ___ ___ / _ \ _  _ __ _ __ _ _| |__ 
| _|| ' \ V / | (_) | || / _' / _' | / /
|___|_||_\_/   \__\_\\_,_\__,_\__,_|_\_\
                                        
Environment Variable Drift Detective 🦆
`
}

// GetSyncMessage returns a message for sync operations
func GetSyncMessage() string {
	return `   __
<(~ )___   Syncing...
 ( ._> /
  '---'`
}
