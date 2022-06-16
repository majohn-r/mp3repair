package internal

// informational log messages
const (
	LOG_BEGIN_EXECUTION          = "execution starts"
	LOG_END_EXECUTION            = "execution ends"
	LOG_EXECUTING_COMMAND        = "executing command"
	LOG_FILE_DELETED             = "successfully deleted file"
	LOG_FILTERING_FILES          = "filtering music files"
	LOG_PARAMETERS_OVERRIDDEN    = "one or more flags were overridden"
	LOG_READING_FILTERED_FILES   = "reading filtered music files"
	LOG_READING_UNFILTERED_FILES = "reading unfiltered music files"
)

// warning log messages
const (
	LOG_CANNOT_COPY_FILE            = "error copying file"
	LOG_CANNOT_CREATE_DIRECTORY     = "cannot create directory"
	LOG_CANNOT_DELETE_DIRECTORY     = "cannot delete directory"
	LOG_CANNOT_DELETE_FILE          = "cannot delete file"
	LOG_CANNOT_EDIT_TRACK           = "cannot edit track"
	LOG_CANNOT_READ_DIRECTORY       = "cannot read directory"
	LOG_CANNOT_READ_FILE            = "cannot read file"
	LOG_CANNOT_UNMARSHAL_YAML       = "cannot unmarshal yaml content"
	LOG_GARBLED_EXTENSION           = "the file extension cannot be parsed as a regular expression"
	LOG_GARBLED_FILTER              = "the filter cannot be parsed as a regular expression"
	LOG_INVALID_EXTENSION_FORMAT    = "the file extension must begin with '.' and contain no other '.' characters"
	LOG_INVALID_FLAG_SETTING        = "flag value is not valid"
	LOG_INVALID_FRAME_VALUE         = "invalid frame value"
	LOG_INVALID_TRACK_NAME          = "the track name cannot be parsed"
	LOG_NO_ARTIST_DIRECTORIES       = "cannot find any artist directories"
	LOG_UNRECOGNIZED_COMMAND        = "unrecognized command"
	LOG_NOT_A_DIRECTORY             = "the file is not a directory"
	LOG_NOTHING_TO_DO               = "the user disabled all functionality"
	LOG_SORTING_OPTION_UNACCEPTABLE = "numeric track sorting is not applicable"
	LOG_UNEXPECTED_VALUE_TYPE       = "unexpected value type"
)

// error log messages
const (
	LOG_COMMANDS_ERROR         = "incorrect number of commands"
	LOG_DEFAULT_COMMANDS_ERROR = "incorrect number of default commands"
)

// error instance messages
const (
	ERROR_NO_APP_DATA_PATH                             = "the APPDATA environment variable is not defined"
	ERROR_NO_DEFAULT_COMMAND_DEFINED                   = "internal error: no subcommand initializers are defined"
	ERROR_NO_SUCH_COMMAND                              = "no command named %q; valid commands include %v"
	ERROR_NO_TEMP_DIRECTORY                            = "neither the TMP nor TEMP environment variables are defined"
	ERROR_INCORRECT_NUMBER_OF_DEFAULT_COMMANDS_DEFINED = "internal error: %d subcommands are designated as default"
)
