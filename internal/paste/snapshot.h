#pragma once

#ifdef __cplusplus
extern "C" {
#endif

/* Snapshots the general pasteboard (all items, all types). Returns the
   number of items captured. The snapshot lives in one module-static slot
   and is replaced by the next call; paste sessions are single-flight. */
int speek_pasteboard_snapshot(void);

/* Writes the last snapshot back to the general pasteboard, replacing its
   current contents. No-op when no snapshot was taken. */
void speek_pasteboard_restore(void);

#ifdef __cplusplus
}
#endif
