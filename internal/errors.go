package internal

// for output to user
const (
	USER_CANNOT_CREATE_DIRECTORY  = "The directory %q cannot be created: %v.\n"
	USER_CANNOT_DELETE_DIRECTORY  = "The directory %q cannot be deleted: %v.\n"
	USER_CANNOT_READ_TOPDIR       = "The -topDir value you specified, %q, cannot be read: %v.\n"
	USER_EXTENSION_INVALID_FORMAT = "The -ext value you specified, %q, must contain exactly one '.' and '.' must be the first character.\n"
	USER_EXTENSION_GARBLED        = "The -ext value you specified, %q, cannot be used for file matching: %v.\n"
	USER_FILTER_GARBLED           = "The %s filter value you specified, %q, cannot be used: %v\n"
	USER_NO_SUCH_COMMAND          = "no command named %q; valid commands include %v"
	USER_SPECIFIED_NO_WORK        = "You disabled all functionality for the command %q.\n"
	USER_TOPDIR_NOT_A_DIRECTORY   = "The -topDir value you specified, %q, is not a directory.\n"
	USER_UNRECOGNIZED_VALUE       = "The %q value you specified, %q, is not valid.\n"
)
