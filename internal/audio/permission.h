#pragma once

// 0 not determined, 1 restricted, 2 denied, 3 authorized. -1 if unavailable.
int speek_mic_status(void);
void speek_mic_request(void);
