Output docs/vhs/compose.gif
Output docs/vhs/testdata/compose.ascii

Set FontSize 14
Set Width 1200
Set Height 600
Set Shell "bash"
Set TypingSpeed 50ms

Env FASTMAIL_ENDPOINT "http://localhost:8080"
Env FASTMAIL_ACCESS_TOKEN "test-token"

# Launch TUI
Type "fastmail-cli tui"
Enter
Sleep 2s

# Open compose
Type "c"
Sleep 1s

# Type To field
Type "alice@example.com"
Tab
Sleep 500ms

# Type Subject
Type "Meeting follow-up"
Tab
Sleep 500ms

# Type Body
Type "Hi Alice, thanks for the great discussion today."
Sleep 1s

# Cancel compose
Escape
Sleep 1s

# Quit
Type "q"
Sleep 500ms
