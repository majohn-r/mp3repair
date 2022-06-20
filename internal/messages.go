package internal

// informational log messages
const (
	LI_BEGIN_EXECUTION          = "execution starts"
	LI_END_EXECUTION            = "execution ends"
	LI_EXECUTING_COMMAND        = "executing command"
	LI_FILE_DELETED             = "successfully deleted file"
	LI_FILTERING_FILES          = "filtering music files"
	LI_NO_SUCH_FILE             = "file does not exist"
	LI_PARAMETERS_OVERRIDDEN    = "one or more flags were overridden"
	LI_READING_FILTERED_FILES   = "reading filtered music files"
	LI_READING_UNFILTERED_FILES = "reading unfiltered music files"
	LI_CONFIGURATION_FILE_READ  = "read configuration file"
)

// warning log messages
const (
	LW_CANNOT_COPY_FILE            = "error copying file"
	LW_CANNOT_CREATE_DIRECTORY     = "cannot create directory"
	LW_CANNOT_DELETE_DIRECTORY     = "cannot delete directory"
	LW_CANNOT_DELETE_FILE          = "cannot delete file"
	LW_CANNOT_EDIT_TRACK           = "cannot edit track"
	LW_CANNOT_READ_DIRECTORY       = "cannot read directory"
	LW_CANNOT_READ_FILE            = "cannot read file"
	LW_CANNOT_UNMARSHAL_YAML       = "cannot unmarshal yaml content"
	LW_GARBLED_EXTENSION           = "the file extension cannot be parsed as a regular expression"
	LW_GARBLED_FILTER              = "the filter cannot be parsed as a regular expression"
	LW_INVALID_EXTENSION_FORMAT    = "the file extension must begin with '.' and contain no other '.' characters"
	LW_INVALID_FLAG_SETTING        = "flag value is not valid"
	LW_INVALID_FRAME_VALUE         = "invalid frame value"
	LW_INVALID_TRACK_NAME          = "the track name cannot be parsed"
	LW_NO_ARTIST_DIRECTORIES       = "cannot find any artist directories"
	LW_UNRECOGNIZED_COMMAND        = "unrecognized command"
	LW_NOT_A_DIRECTORY             = "the file is not a directory"
	LW_NOTHING_TO_DO               = "the user disabled all functionality"
	LW_SORTING_OPTION_UNACCEPTABLE = "numeric track sorting is not applicable"
	LW_UNEXPECTED_VALUE_TYPE       = "unexpected value type"
)

// error log messages
const (
	LE_CANNOT_DETERMINE_FILE_STATUS = "cannot determine file status"
	LE_COMMAND_COUNT                = "incorrect number of commands"
	LE_DEFAULT_COMMAND_COUNT        = "incorrect number of default commands"
	LE_FILE_IS_DIR                  = "file is a directory"
)

// for output to user
const (
	USER_CANNOT_CREATE_DIRECTORY                      = "The directory %q cannot be created: %v.\n"
	USER_CANNOT_DELETE_DIRECTORY                      = "The directory %q cannot be deleted: %v.\n"
	USER_CANNOT_READ_TOPDIR                           = "The -topDir value you specified, %q, cannot be read: %v.\n"
	USER_CONFIGURATION_FILE_GARBLED                   = "The configuration file %q is not well-formed YAML: %v\n"
	USER_CONFIGURATION_FILE_IS_DIR                    = "The configuration file %q is a directory.\n"
	USER_EXTENSION_INVALID_FORMAT                     = "The -ext value you specified, %q, must contain exactly one '.' and '.' must be the first character.\n"
	USER_EXTENSION_GARBLED                            = "The -ext value you specified, %q, cannot be used for file matching: %v.\n"
	USER_FILTER_GARBLED                               = "The %s filter value you specified, %q, cannot be used: %v\n"
	USER_INCORRECT_NUMBER_OF_DEFAULT_COMMANDS_DEFINED = "An internal error has occurred: there are %d default commands!\n"
	USER_INVALID_SORTING_APPLIED                      = "The value of the %s flag, '%s', cannot be used unless '%s' is true; track sorting will be alphabetic.\n"
	USER_NO_APPDATA_FOLDER                            = "The APPDATA environment variable is not defined.\n"
	USER_NO_COMMANDS_DEFINED                          = "An internal error has occurred: no commands are defined!\n"
	USER_NO_SUCH_COMMAND                              = "There is no command named %q; valid commands include %v.\n"
	USER_NO_TEMP_FOLDER                               = "Neither the TMP nor TEMP environment variables are defined.\n"
	USER_SPECIFIED_NO_WORK                            = "You disabled all functionality for the command %q.\n"
	USER_TOPDIR_NOT_A_DIRECTORY                       = "The -topDir value you specified, %q, is not a directory.\n"
	USER_UNRECOGNIZED_VALUE                           = "The %q value you specified, %q, is not valid.\n"
)

const (
	ERROR_DIR_IS_FILE               = "file exists and is not a directory"
	ERROR_DOES_NOT_BEGIN_WITH_DIGIT = "first character is not a digit"
	ERROR_EDIT_UNNECESSARY          = "no edit required"
	ERROR_FILE_IS_DIR               = "file exists but is a directory"
	ERROR_ZERO_LENGTH               = "zero length"
)
