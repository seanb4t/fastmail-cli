Output docs/vhs/dashboard.gif
Output docs/vhs/testdata/dashboard.ascii

Set FontSize 14
Set Width 1200
Set Height 600
Set Shell "bash"
Set TypingSpeed 50ms

# Set environment for mock Fastmail
Env FASTMAIL_ENDPOINT "http://localhost:8080"
Env FASTMAIL_ACCESS_TOKEN "test-token"

# Launch TUI
Type "fastmail-cli tui"
Enter
Sleep 2s

# Browse — press down to highlight second email
Down
Sleep 1s

# Read first email
Enter
Sleep 2s

# Go back
Escape
Sleep 1s

# Quit
Type "q"
Sleep 500ms
