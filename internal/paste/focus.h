#pragma once

#ifdef __cplusplus
extern "C" {
#endif

int speek_frontmost_pid(void);
int speek_current_pid(void);
int speek_activate_pid(int pid);

#ifdef __cplusplus
}
#endif
