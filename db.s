#include "textflag.h"

TEXT ·DBString(SB), NOSPLIT, $-16
	MOVQ	ptr+0(FP), AX
	MOVQ	AX, ret_base+16(FP)

	MOVQ	len+8(FP), AX
	MOVQ	AX, ret_len+24(FP)
	RET

TEXT ·DBSlice(SB), NOSPLIT, $-16
	MOVQ	ptr+0(FP), AX
	MOVQ	AX, ret_base+16(FP)

	MOVQ	len+8(FP), AX
	MOVQ	AX, ret_len+24(FP)

	MOVQ	len+8(FP), AX
	MOVQ	AX, ret_cap+32(FP)
	RET
