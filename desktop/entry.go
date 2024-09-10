package desktop

// Entry presents a Desktop Entry specified by the [Desktop Entry Specification] version 1.5.
//
// [Desktop Entry Specification]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/
type Entry struct {

	// Type of the desktop entry. The specification defines 3 types of desktop entries:
	//  - Application (type 1),
	//  - Link (type 2)
	//  - Directory (type 3).
	// To allow the addition of new types in the future, implementations should ignore desktop
	// entries with an unknown type.
	Type string

	// Version of the Desktop Entry Specification that the desktop entry conforms
	// with. Note that the version field is not required to be present.
	Version string

	// Name is the specific name of the application, for example "Firefox".
	Name LocaleString

	// GenericName is a generic name of the application, for example "Web Browser".
	GenericName LocaleString

	// NoDisplay means "this application exists, but don't display it in the menus".
	// This can be useful to e.g. associate this application with MIME types, so that
	// it gets launched from a file manager (or other apps), without having a menu
	// entry for it (there are tons of good reasons for this, including e.g. the
	// netscape -remote, or kfmclient openURL kind of stuff).
	NoDisplay bool

	// Comment is the tooltip for the entry, for example "View sites on the Internet".
	// The value should not be redundant with the values of Name and GenericName.
	Comment LocaleString

	// Icon to display in file manager, menus, etc. If the name is an absolute path,
	// the given file will be used. If the name is not an absolute path, the
	// algorithm described in the [Icon Theme Specification] will be used to locate the
	// icon.
	//
	// [Icon Theme Specification]: http://freedesktop.org/wiki/Standards/icon-theme-spec
	Icon IconString

	// Hidden should have been called Deleted. It means the user deleted (at their
	// level) something that was present (at an upper level, e.g. in the system
	// dirs). It's strictly equivalent to the .desktop file not existing at all, as
	// far as that user is concerned. This can also be used to "uninstall" existing
	// files (e.g. due to a renaming) - by letting `make install` install a file with
	// Hidden=true in it.
	Hidden bool

	// A list of strings identifying the desktop environments that should display/not
	// display a given desktop entry.
	//
	// By default, a desktop file should be shown, unless an OnlyShowIn key is
	// present, in which case, the default is for the file not to be shown.
	//
	// If $XDG_CURRENT_DESKTOP is set then it contains a colon-separated list of
	// strings. In order, each string is considered. If a matching entry is found in
	// OnlyShowIn then the desktop file is shown. If an entry is found in NotShowIn
	// then the desktop file is not shown. If none of the strings match then the
	// default action is taken (as above).
	//
	// $XDG_CURRENT_DESKTOP should have been set by the login manager, according to
	// the value of the DesktopNames found in the session file. The entry in the
	// session file has multiple values separated in the usual way: with a semicolon.
	//
	// The same desktop name may not appear in both OnlyShowIn and NotShowIn of a group.
	OnlyShowIn []string

	// NotShowIn is the opposite of OnlyShowIn, see that field for explanation.
	NotShowIn []string

	// DBusActivatable specifies if D-Bus activation is supported for this
	// application. If this key is missing, the default value is false. If the value
	// is true then implementations should ignore the Exec key and send a D-Bus
	// message to launch the application. See D-Bus Activation for more information
	// on how this works. Applications should still include Exec= lines in their
	// desktop files for compatibility with implementations that do not understand
	// the DBusActivatable key.
	DBusActivatable bool

	// TryExec is a path to an executable file on disk used to determine if the program is
	// actually installed. If the path is not an absolute path, the file is looked up
	// in the $PATH environment variable. If the file is not present or if it is not
	// executable, the entry may be ignored (not be used in menus, for example).
	TryExec string

	// Exec defines the program to execute, possibly with arguments. See the Exec key for details on
	// how this key works. The Exec key is required if DBusActivatable is not set to
	// true.
	//
	// Even if DBusActivatable is true, Exec should be specified for compatibility with
	// implementations that do not understand DBusActivatable.
	//
	// Specified at [The Exec key].
	//
	// [The Exec key]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/exec-variables.html
	Exec ExecValue

	// If entry is of type Application, the working directory to run the program in.
	Path string

	// Whether the program runs in a terminal window.
	Terminal bool

	// Actions contains identifiers for application actions.
	// This can be used to tell the application to make a specific action, different from the
	// default behavior.
	// This is specified in [Additional applications actions].
	//
	// [Additional applications actions]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/extra-actions.html
	Actions []Action

	// The MIME type(s) supported by this application.
	MimeType []string

	// Categories in which the entry should be shown in a menu (for possible values see the
	// [Desktop Menu Specification]).
	//
	// [Desktop Menu Specification]: http://www.freedesktop.org/Standards/menu-spec
	Categories []string

	// Implements contains a list of interfaces that this application implements.
	// By default, a desktop file implements no interfaces.
	// See [Interfaces] for more information on how this works.
	//
	// [Interfaces]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/interfaces.html
	Implements []string

	// Keywords is a list of strings which may be used in addition to other metadata to describe
	// this entry.
	// This can be useful e.g. to facilitate searching through entries.
	// The values are not meant for display, and should not be redundant with the values of Name or
	// GenericName.
	Keywords LocaleStrings

	// StartupNotify determines support for startup notifications.
	//  - absent (0): Support is unknown. Implementations can choose how they want to handle it.
	//  - true (1): it is KNOWN that the application will send a "remove" message when started with
	//    the DESKTOP_STARTUP_ID environment variable set.
	//  - false (2): it is KNOWN that the application does not work with startup notification at
	//    all. This can be because it does not show any window, breaks even when using
	//    StartupWMClass, ...
	//
	// See the [Startup Notification Protocol Specification] for more details.
	//
	// [Startup Notification Protocol Specification]: http://www.freedesktop.org/Standards/startup-notification-spec
	StartupNotify int

	// StartupWMClass, if specified, it is known that the application will map at
	// least one window with the given string as its WM class or WM name hint (see
	// the [Startup Notification Protocol Specification] for more details).
	//
	// [Startup Notification Protocol Specification]: http://www.freedesktop.org/Standards/startup-notification-spec
	StartupWMClass string

	// URL is present on Type == Link.
	URL string

	// PrefersNonDefaultGPU, if true, signals that the application prefers to be run on a more
	// powerful discrete GPU if available, which we describe as “a GPU other than the default one”
	// in this spec to avoid the need to define what a discrete GPU is and in which cases it
	// might be considered more powerful than the default GPU.
	// This key is only a hint and support might not be present depending on the implementation.
	PrefersNonDefaultGPU bool

	// SingleMainWindow signals that, if true, the application has a single main
	// window, and does not support having an additional one opened. This key is used
	// to signal to the implementation to avoid offering a UI to launch another
	// window of the app.
	// This key is only a hint and support might not be present depending on the implementation.
	SingleMainWindow bool

	// OtherKeys is a map of the remaining keys in the "Desktop Entry" group.
	OtherKeys map[string]string

	// OtherGroups holds the data of groups other than the "Desktop Entry" group in the desktop
	// file.
	// The format is Key=Group name, Value=Map of key-value pairs.
	OtherGroups map[string]map[string]string
}

type Action struct {

	// Name contains the label that will be shown to the user. Since actions are
	// always shown in the context of a specific application (that is, as a submenu
	// of a launcher), this only needs to be unambiguous within one application and
	// should not include the application name.
	Name LocaleString

	// Icon to be shown together with the action.
	// If the name is an absolute path, the given file will be used.
	// If the name is not an absolute path, the
	// algorithm described in the [Icon Theme Specification] will be used to locate
	// the icon.
	// Implementations may choose to ignore it.
	//
	// [Icon Theme Specification]: http://freedesktop.org/wiki/Standards/icon-theme-spec
	Icon IconString

	// Exec contains the program to execute for this action, possibly with arguments.
	// See the [Exec key] for details on how this key works.
	// The Exec key is required if
	// DBusActivatable is not set to true in the main desktop entry group. Even if
	// DBusActivatable is true, Exec should be specified for compatibility with
	// implementations that do not understand DBusActivatable.
	//
	// [Exec key]: https://specifications.freedesktop.org/desktop-entry-spec/1.5/exec-variables.html
	Exec ExecValue
}
