#!/bin/bash

# Name of the tmux session
SESSION_NAME="srbft"

# Start a new tmux session
tmux kill-session -t $SESSION_NAME
tmux new-session -d -s $SESSION_NAME

# Split the window into two panes horizontally
tmux split-window -h

# Split the first pane into two vertically
tmux select-pane -t 0
tmux split-window -v

# Split the second pane (which is now the third pane) into two vertically
tmux select-pane -t 2
tmux split-window -v

# Pane 0: Launch command for node 1
tmux send-keys -t ${SESSION_NAME}:0.0 './sr-bft pbft node -id 0' C-m

# Pane 1: Launch command for node 2
tmux send-keys -t ${SESSION_NAME}:0.1 './sr-bft pbft node -id 1' C-m

# Pane 2: Launch command for node 3
tmux send-keys -t ${SESSION_NAME}:0.2 './sr-bft pbft node -id 2' C-m

# Pane 3: Launch command for node 4
tmux send-keys -t ${SESSION_NAME}:0.3 './sr-bft pbft node -id 3' C-m

# Attach to the session
tmux attach-session -t $SESSION_NAME
