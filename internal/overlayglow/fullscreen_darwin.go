//go:build darwin

package overlayglow

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

static NSRect speekUnconstrained(id self, SEL cmd, NSRect frameRect, NSScreen *screen) {
	(void)self;
	(void)cmd;
	(void)screen;
	return frameRect;
}

static void speekAllowFullFrame(NSWindow *w) {
	Class cls = object_getClass(w);
	SEL sel = @selector(constrainFrameRect:toScreen:);
	const char *types = "{CGRect={CGPoint=dd}{CGSize=dd}}@:{CGRect={CGPoint=dd}{CGSize=dd}}@";
	if (class_addMethod(cls, sel, (IMP)speekUnconstrained, types)) {
		return;
	}
	Method m = class_getInstanceMethod(cls, sel);
	if (m != NULL) {
		method_setImplementation(m, (IMP)speekUnconstrained);
	}
}

static NSScreen *speekScreenForCenter(double cx, double cy) {
	NSPoint center = NSMakePoint(cx, cy);
	for (NSScreen *s in [NSScreen screens]) {
		if (NSPointInRect(center, s.frame)) {
			return s;
		}
	}
	return [NSScreen mainScreen];
}

static void speekPlaceOnScreen(NSWindow *w, NSScreen *screen) {
	if (w == nil || screen == nil) {
		return;
	}
	speekAllowFullFrame(w);
	w.level = NSScreenSaverWindowLevel;
	w.styleMask |= NSWindowStyleMaskFullSizeContentView;
	w.titlebarAppearsTransparent = YES;
	w.titleVisibility = NSWindowTitleHidden;
	w.collectionBehavior |= NSWindowCollectionBehaviorCanJoinAllSpaces | NSWindowCollectionBehaviorFullScreenAuxiliary;
	[w setFrame:screen.frame display:YES];
	[w setHidesOnDeactivate:NO];

	NSView *content = w.contentView;
	NSView *frameView = content.superview;
	if (frameView != nil) {
		content.frame = frameView.bounds;
	}
	NSRect bounds = content.bounds;
	for (NSView *sub in content.subviews) {
		sub.frame = bounds;
	}
}

void speekPlaceOverDisplay(void *window) {
	NSWindow *w = (NSWindow *)window;
	if (w == nil) {
		return;
	}
	NSPoint p = [NSEvent mouseLocation];
	speekPlaceOnScreen(w, speekScreenForCenter(p.x, p.y));
}

void speekPlaceOverFrame(void *window, double x, double y, double width, double height) {
	NSWindow *w = (NSWindow *)window;
	if (w == nil) {
		return;
	}
	double cx = x + width * 0.5;
	double cy = y + height * 0.5;
	speekPlaceOnScreen(w, speekScreenForCenter(cx, cy));
}
*/
import "C"
import "unsafe"

// PlaceOverDisplay sizes the native window to the screen under the pointer,
// including the menu bar and Dock, and raises it above both.
func PlaceOverDisplay(native uintptr) {
	if native == 0 {
		return
	}
	C.speekPlaceOverDisplay(unsafe.Pointer(native))
}

// PlaceOverFrame places the window on the display whose bounds contain the
// frame center (locked at recording start).
func PlaceOverFrame(native uintptr, x, y, width, height float64) {
	if native == 0 {
		return
	}
	C.speekPlaceOverFrame(unsafe.Pointer(native), C.double(x), C.double(y), C.double(width), C.double(height))
}
