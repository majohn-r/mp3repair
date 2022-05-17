package internal

// for logging
const (
	LOG_ENV_ISSUES_DETECTED               string = "at least one required environment variable is not set"
	LOG_NO_APP_DATA_PATH                  string = "environment: APPDATA is not defined"
	LOG_NO_HOME_PATH                      string = "environment: HOMEPATH is not defined"
	LOG_NO_TEMP_DIRECTORY                 string = "environment: neither TMP nor TEMP are defined"
	LOG_NO_DEFAULT_COMMAND_DEFINED        string = "internal error: no subcommand initializers defined"
	LOG_TOO_MANY_DEFAULT_COMMANDS_DEFINED string = "internal error: only 1 subcommand should be designated as default; %d were found"
	LOG_CANNOT_DELETE_FILE                string = "cannot delete file"
	LOG_CANNOT_READ_DIRECTORY             string = "cannot read directory"
	LOG_CANNOT_READ_FILE                  string = "cannot read file"
	LOG_INVALID_TRACK_NAME                string = "the track name does not conform to the expected syntax"
	LOG_NO_ARTIST_DIRECTORIES             string = "cannot find any artist directories"
	LOG_GARBLED_EXTENSION                 string = "the file extension cannot be used for file matching"
	LOG_GARBLED_FILTER                    string = "the filter cannot be used"
	LOG_INVALID_EXTENSION_FORMAT          string = "the file extension must contain exactly one '.' and '.' must be the first character"
	LOG_INVALID_FLAG_SETTING              string = "the command flag value is not valid"
	LOG_NO_SUCH_COMMAND                   string = "no command named %q; valid commands include %v"
	LOG_NOT_A_DIRECTORY                   string = "the specified file is not a directory"
	LOG_NOTHING_TO_DO                     string = "the user disabled all functionality"
)

// for output to user
const (
	USER_CANNOT_CREATE_DIRECTORY  string = "The directory %q cannot be created: %v.\n"
	USER_CANNOT_READ_TOPDIR       string = "The -topDir value you specified, %q, cannot be read: %v.\n"
	USER_EXTENSION_INVALID_FORMAT string = "The -ext value you specified, %q, must contain exactly one '.' and '.' must be the first character.\n"
	USER_EXTENSION_GARBLED        string = "The -ext value you specified, %q, cannot be used for file matching: %v.\n"
	USER_FILTER_GARBLED           string = "The %s filter value you specified, %q, cannot be used: %v\n"
	USER_SPECIFIED_NO_WORK        string = "You disabled all functionality for the command %q.\n"
	USER_TOPDIR_NOT_A_DIRECTORY   string = "The -topDir value you specified, %q, is not a directory.\n"
	USER_UNRECOGNIZED_VALUE       string = "The %q value you specified, %q, is not valid.\n"
)
