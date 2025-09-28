[![progress-banner](https://backend.codecrafters.io/progress/shell/4e62f1ba-68a9-4bee-9cc2-bc22c187210e)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

# My POSIX-Compliant Shell

This is my custom implementation of a **POSIX-compliant shell**, built as a personal project to deeply understand how shells work and experiment with performance optimizations.  

While it doesn't support all the features from the unix shell it does support big part of it. Like:

- Parse and interpret shell commands  
- Execute **external programs**  
- Implement **builtin commands** such as `cd`, `pwd`, `echo`, `exit`, and `history`  
- Integrate **GNU Readline** for command line editing, history, and tab-completion  
- Handle **in-memory command history** with optional persistence to a history file, including `HISTSIZE` and `HISTFILESIZE`  

Although many of the optimizations implemented are **not strictly necessary**, this project served as a playground to **push the boundaries of shell performance** and explore low-level memory handling, efficient data structures, game the GOLANG escape analysis and Go-C interop via CGO.  

The shell is designed to behave similarly to Bash:

- Commands are stored in memory as they are entered.  
- The history file (`HISTFILE`) is only updated when the shell exits or when explicitly requested.  
- Users can also read, write, or append history to arbitrary files while keeping per-file tracking.  

**Note**: This project is primarily for learning purposes. For a structured learning experience on building a shell, check out [codecrafters.io](https://codecrafters.io).  
