#import <AppKit/AppKit.h>
#include <stdlib.h>
#include "snapshot.h"

// One module-static slot: snapshot -> restore is used once per paste session,
// and the coordinator guarantees a single paste at a time.
static NSMutableArray *g_snapshot = nil;

int speek_pasteboard_snapshot(void) {
    @autoreleasepool {
        [g_snapshot release];
        g_snapshot = [NSMutableArray new];
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        NSArray *items = [pb pasteboardItems];
        if (items == nil) {
            return 0;
        }
        for (NSPasteboardItem *item in items) {
            NSMutableDictionary *data = [NSMutableDictionary dictionary];
            for (NSString *type in [item types]) {
                NSData *value = [item dataForType:type];
                if (value != nil) {
                    data[type] = value;
                }
            }
            [g_snapshot addObject:data];
        }
        return (int)[g_snapshot count];
    }
}

void speek_pasteboard_restore(void) {
    @autoreleasepool {
        if (g_snapshot == nil) {
            return;
        }
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        [pb clearContents];
        if ([g_snapshot count] == 0) {
            return;
        }
        NSMutableArray *items = [NSMutableArray array];
        for (NSDictionary *data in g_snapshot) {
            NSPasteboardItem *item = [NSPasteboardItem new];
            for (NSString *type in data) {
                [item setData:data[type] forType:type];
            }
            [items addObject:item];
            [item release];
        }
        [pb writeObjects:items];
    }
}
