#import <AppKit/AppKit.h>
#include "focus.h"

int speek_frontmost_pid(void) {
    NSRunningApplication *app = [[NSWorkspace sharedWorkspace] frontmostApplication];
    if (app == nil) {
        return -1;
    }
    return (int)[app processIdentifier];
}

int speek_current_pid(void) {
    return (int)[[NSRunningApplication currentApplication] processIdentifier];
}

int speek_activate_pid(int pid) {
    if (pid <= 0) {
        return 0;
    }
    for (NSRunningApplication *app in [[NSWorkspace sharedWorkspace] runningApplications]) {
        if ((int)[app processIdentifier] == pid) {
            return [app activateWithOptions:NSApplicationActivateIgnoringOtherApps] ? 1 : 0;
        }
    }
    return 0;
}
