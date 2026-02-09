Output docs/vhs/navigation.gif
Output docs/vhs/testdata/navigation.ascii

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

# Cycle panes with tab
Tab
Sleep 1s
Tab
Sleep 1s
Tab
Sleep 1s

# Toggle sidebar
Type "b"
Sleep 1s

# Toggle sidebar back
Type "b"
Sleep 1s

# Navigate mailbox list
Tab
Down
Down
Enter
Sleep 2s

# Quit
Type "q"
Sleep 500ms
