#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dirent.h>
#include <sys/stat.h>
#include <unistd.h>
#include <readline/readline.h>
#include <readline/history.h>

// Just a large enough number
#define MAX_MATCHES 4096

// Custom generator for a fixed set of completions
char* commands[] = {"help", "exit", "run", "status", "type", "history",NULL};

char* matches[MAX_MATCHES];
int match_count = 0;


int is_executable_by_user(const char* path) {
    return access(path, X_OK) == 0;
}

// This function populates the global 'matches' array with command names
// that match the beginning of 'text'. It searches:
//   1. Built-in commands defined in 'commands[]'
//   2. Executable files found in directories from the PATH environment variable
//
// NOTE:
// - Each match is stored using strdup(), which allocates memory on the heap.
// - You MUST free each match later using `free(matches[i])` when you're done using them,
//   unless you are passing them to `rl_completion_matches()`, which will handle cleanup.
void collect_command_matches(const char* text) {
    match_count = 0;
    int len = strlen(text);  // Store length of input for prefix comparisons

    // Step 1: Match built-in commands
    for (int i = 0; commands[i] && match_count < MAX_MATCHES; i++) {
        // strncmp compares the first 'len' characters of both strings
        // If text is "he", this matches "help", "hello", etc.
        if (strncmp(commands[i], text, len) == 0) {
            // strdup creates a heap-allocated copy of the string
            // This is needed because readline may store and use the string after this function ends
            matches[match_count++] = strdup(commands[i]);
        }
    }

    // Step 2: Match executables in $PATH
    char* path = getenv("PATH");  // Get the PATH environment variable
    if (!path) return;            // No PATH? Exit early

    // Create a copy of PATH because strtok modifies the string
    char* path_dup = strdup(path);
    char* dir = strtok(path_dup, ":");  // Split PATH by ':'

    while (dir && match_count < MAX_MATCHES) {
        DIR* d = opendir(dir);  // Open the directory (e.g., "/usr/bin")
        if (!d) {
			//Passing NULL to strtok() tells him to keep splitting the same string it was working on previously.
			//Internally, strtok() maintains a static pointer to remember where it left off.
            dir = strtok(NULL, ":");  // Move to the next PATH entry
            continue;
        }

        struct dirent* entry;
        while ((entry = readdir(d)) && match_count < MAX_MATCHES) {
            // Only consider regular files, symlinks, or unknown types
			// -> is used to deference a file
            if (entry->d_type != DT_REG &&
                entry->d_type != DT_LNK &&
                entry->d_type != DT_UNKNOWN)
                continue;

            // Check if the file name starts with the user input
            if (strncmp(entry->d_name, text, len) == 0) {
				// Construct the full absolute path to the potential executable file.
				// Example: if dir = "/usr/bin" and entry->d_name = "python",
				//          then fullpath will be "/usr/bin/python".
				// snprintf writes into 'fullpath' safely, preventing buffer overflows
				// by limiting the output to the size of the array (1024 bytes).
                char fullpath[1024];
                snprintf(fullpath, sizeof(fullpath), "%s/%s", dir, entry->d_name);

                // Check if this file is executable by the user(permissions)
                if (is_executable_by_user(fullpath)) {
                    // strdup again to make a persistent copy (heap-allocated)
                    matches[match_count++] = strdup(entry->d_name);
                }
            }
        }

        closedir(d);               // Clean up directory stream
		//Passing NULL to strtok() tells him to keep splitting the same string it was working on previously.
		//Internally, strtok() maintains a static pointer to remember where it left off.
        dir = strtok(NULL, ":");   // Move to next directory in PATH
    }

    free(path_dup);  // Free duplicated PATH string

    // ❗️NOTE: Remember to free each matches[i] later unless you're passing the array to rl_completion_matches()
}

// === Filename autocomplete ===
char* filename_generator(const char* text, int state) {
	static DIR* dir = NULL;
	static struct dirent* entry;
	static char dirname[1024];
	static int len;

	if (state == 0) {
		char* slash = strrchr(text, '/');
		if (slash) {
			size_t pathlen = slash - text + 1;
			strncpy(dirname, text, pathlen);
			dirname[pathlen] = '\0';
			dir = opendir(dirname);
			text = slash + 1;
		} else {
			strcpy(dirname, "./");
			dir = opendir(dirname);
		}
		len = strlen(text);
	}

	if (!dir) return NULL;

	while ((entry = readdir(dir))) {
		if (strncmp(entry->d_name, text, len) == 0) {
			static char fullpath[1024];
			snprintf(fullpath, sizeof(fullpath), "%s%s", dirname, entry->d_name);
			return strdup(fullpath);
		}
	}

	closedir(dir);
	dir = NULL;
	return NULL;
}

// command_generator is called repeatedly by readline to generate autocomplete suggestions.
// The `state` parameter works like this:
//   - state == 0: readline is starting a new completion cycle. We use this to initialize
//                 our internal state by collecting all possible matches (e.g., built-ins, $PATH).
//   - state > 0: readline is asking for the next match. We return matches[state - 1] each time.
//
// readline will keep calling this function with incrementing state values until we return NULL,
// indicating that there are no more completions left.
char* command_generator(const char* text, int state) {
    if (state == 0) {
        collect_command_matches(text);  // Populate the global matches[] array
    }

    if (state < match_count) {
        return matches[state];  // Return one match per call
    }

    return NULL;  // No more matches
}

char** my_completion(const char* text, int start, int end) {
	if (start == 0) {
		return rl_completion_matches(text, command_generator);
	} else {
		return rl_completion_matches(text, filename_generator);
	}
}

// Set the completion function
void setup_completion() {
    rl_attempted_completion_function = my_completion;
}