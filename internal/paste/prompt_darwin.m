#import <ApplicationServices/ApplicationServices.h>
#include "prompt.h"

int speek_prompt_accessibility(void) {
    const void *keys[] = {kAXTrustedCheckOptionPrompt};
    const void *values[] = {kCFBooleanTrue};
    CFDictionaryRef options = CFDictionaryCreate(NULL, keys, values, 1,
        &kCFTypeDictionaryKeyCallBacks, &kCFTypeDictionaryValueCallBacks);
    int trusted = AXIsProcessTrustedWithOptions(options) ? 1 : 0;
    CFRelease(options);
    return trusted;
}
