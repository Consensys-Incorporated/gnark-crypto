//go:build !purego
#include "textflag.h"
TEXT ·permutation16x16x512_arm64(SB), $128-40
	MOVD matrix+0(FP), R0
	MOVD roundKeys+8(FP), R1
	MOVD result+32(FP), R2
	MOVD $0x7f000001, R3
	MOVD $0x7effffff, R4
	VDUP R3, V0.S4
	VDUP R4, V1.S4
	MOVD $1, R5
	VDUP R5, V28.S4

#define ADD_MOD(in0, in1, in2) \
	VADD  in0.S4, in1.S4, V30.S4 \
	VSUB  V0.S4, V30.S4, V31.S4  \
	VUMIN V30.S4, V31.S4, in2.S4 \

#define SUB_MOD(in0, in1, in2) \
	VSUB  in1.S4, in0.S4, V30.S4 \
	VADD  V0.S4, V30.S4, V31.S4  \
	VUMIN V30.S4, V31.S4, in2.S4 \

#define DOUBLE_MOD(in0, in1) \
	VSHL  $1, in0.S4, V30.S4     \
	VSUB  V0.S4, V30.S4, V31.S4  \
	VUMIN V30.S4, V31.S4, in1.S4 \

#define TRIPLE_MOD(in0, in1) \
DOUBLE_MOD(in0, V30)   \
ADD_MOD(V30, in0, in1) \

#define QUAD_MOD(in0, in1) \
DOUBLE_MOD(in0, V30) \
DOUBLE_MOD(V30, in1) \

#define MAT_MUL_4(in0, in1, in2, in3) \
	ADD_MOD(in0, in1, V18) \
	ADD_MOD(in2, in3, V19) \
	ADD_MOD(V18, V19, V20) \
	ADD_MOD(V20, in1, V21) \
	ADD_MOD(V20, in3, V22) \
	DOUBLE_MOD(in0, in3)   \
	ADD_MOD(in3, V22, in3) \
	DOUBLE_MOD(in2, in1)   \
	ADD_MOD(in1, V21, in1) \
	ADD_MOD(V18, V21, in0) \
	ADD_MOD(V19, V22, in2) \

#define MAT_MUL_EXT(in0, in1, in2, in3, in4, in5, in6, in7, in8, in9, in10, in11, in12, in13, in14, in15) \
MAT_MUL_4(in0, in1, in2, in3)     \
MAT_MUL_4(in4, in5, in6, in7)     \
MAT_MUL_4(in8, in9, in10, in11)   \
MAT_MUL_4(in12, in13, in14, in15) \
ADD_MOD(in0, in4, V18)            \
ADD_MOD(V18, in8, V18)            \
ADD_MOD(V18, in12, V18)           \
ADD_MOD(in1, in5, V19)            \
ADD_MOD(V19, in9, V19)            \
ADD_MOD(V19, in13, V19)           \
ADD_MOD(in2, in6, V20)            \
ADD_MOD(V20, in10, V20)           \
ADD_MOD(V20, in14, V20)           \
ADD_MOD(in3, in7, V21)            \
ADD_MOD(V21, in11, V21)           \
ADD_MOD(V21, in15, V21)           \
ADD_MOD(in0, V18, in0)            \
ADD_MOD(in1, V19, in1)            \
ADD_MOD(in2, V20, in2)            \
ADD_MOD(in3, V21, in3)            \
ADD_MOD(in4, V18, in4)            \
ADD_MOD(in5, V19, in5)            \
ADD_MOD(in6, V20, in6)            \
ADD_MOD(in7, V21, in7)            \
ADD_MOD(in8, V18, in8)            \
ADD_MOD(in9, V19, in9)            \
ADD_MOD(in10, V20, in10)          \
ADD_MOD(in11, V21, in11)          \
ADD_MOD(in12, V18, in12)          \
ADD_MOD(in13, V19, in13)          \
ADD_MOD(in14, V20, in14)          \
ADD_MOD(in15, V21, in15)          \

	MOVD $0, R7

batch_loop:
	VEOR V2.B16, V2.B16, V2.B16
	VEOR V3.B16, V3.B16, V3.B16
	VEOR V4.B16, V4.B16, V4.B16
	VEOR V5.B16, V5.B16, V5.B16
	VEOR V6.B16, V6.B16, V6.B16
	VEOR V7.B16, V7.B16, V7.B16
	VEOR V8.B16, V8.B16, V8.B16
	VEOR V9.B16, V9.B16, V9.B16
	VEOR V10.B16, V10.B16, V10.B16
	VEOR V11.B16, V11.B16, V11.B16
	VEOR V12.B16, V12.B16, V12.B16
	VEOR V13.B16, V13.B16, V13.B16
	VEOR V14.B16, V14.B16, V14.B16
	VEOR V15.B16, V15.B16, V15.B16
	VEOR V16.B16, V16.B16, V16.B16
	VEOR V17.B16, V17.B16, V17.B16
	MOVD $0, R8
	LSL  $13, R7, R13
	ADD  R0, R13, R9
	ADD  $0x800, R9, R10
	ADD  $0x800, R10, R11
	ADD  $0x800, R11, R12

step_loop:
	MOVWU  (R9), R13
	VMOV   R13, V18.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V18.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V18.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V18.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	MOVWU  (R9), R13
	VMOV   R13, V19.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V19.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V19.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V19.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	MOVWU  (R9), R13
	VMOV   R13, V20.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V20.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V20.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V20.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	MOVWU  (R9), R13
	VMOV   R13, V21.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V21.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V21.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V21.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	MOVWU  (R9), R13
	VMOV   R13, V22.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V22.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V22.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V22.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	MOVWU  (R9), R13
	VMOV   R13, V23.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V23.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V23.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V23.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	MOVWU  (R9), R13
	VMOV   R13, V24.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V24.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V24.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V24.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	MOVWU  (R9), R13
	VMOV   R13, V25.S[0]
	MOVWU  (R10), R13
	VMOV   R13, V25.S[1]
	MOVWU  (R11), R13
	VMOV   R13, V25.S[2]
	MOVWU  (R12), R13
	VMOV   R13, V25.S[3]
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V18.B16, V10.B16
	VMOV   V19.B16, V11.B16
	VMOV   V20.B16, V12.B16
	VMOV   V21.B16, V13.B16
	VMOV   V22.B16, V14.B16
	VMOV   V23.B16, V15.B16
	VMOV   V24.B16, V16.B16
	VMOV   V25.B16, V17.B16
	MOVD   RSP, R13
	VST1.P [V18.S4], 16(R13)
	VST1.P [V19.S4], 16(R13)
	VST1.P [V20.S4], 16(R13)
	VST1.P [V21.S4], 16(R13)
	VST1.P [V22.S4], 16(R13)
	VST1.P [V23.S4], 16(R13)
	VST1.P [V24.S4], 16(R13)
	VST1.P [V25.S4], 16(R13)
	MAT_MUL_EXT(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   0(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   0(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   0(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   0(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   0(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   0(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   0(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   0(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   0(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   0(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   0(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   0(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   0(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   0(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   0(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   0(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   24(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   24(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   24(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   24(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   24(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   24(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   24(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   24(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   24(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   24(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   24(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   24(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   24(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   24(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   24(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   24(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   48(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   48(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   48(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   48(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   48(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   48(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   48(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   48(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   48(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   48(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   48(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   48(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   48(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   48(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   48(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   48(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   72(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   96(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   120(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   144(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   168(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   192(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   216(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   240(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   264(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   288(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   312(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   336(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   360(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   384(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   408(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   432(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   456(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   480(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   504(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   528(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   552(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD(V2, V3, V18)
	ADD_MOD(V4, V5, V19)
	ADD_MOD(V6, V7, V20)
	ADD_MOD(V8, V9, V21)
	ADD_MOD(V18, V19, V18)
	ADD_MOD(V20, V21, V20)
	ADD_MOD(V18, V20, V18)
	ADD_MOD(V10, V11, V22)
	ADD_MOD(V12, V13, V23)
	ADD_MOD(V14, V15, V24)
	ADD_MOD(V16, V17, V25)
	ADD_MOD(V22, V23, V22)
	ADD_MOD(V24, V25, V24)
	ADD_MOD(V22, V24, V22)
	ADD_MOD(V18, V22, V18)
	DOUBLE_MOD(V2, V2)
	SUB_MOD(V18, V2, V2)
	ADD_MOD(V18, V3, V3)
	DOUBLE_MOD(V4, V4)
	ADD_MOD(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD(V18, V5, V5)
	TRIPLE_MOD(V6, V6)
	ADD_MOD(V18, V6, V6)
	QUAD_MOD(V7, V7)
	ADD_MOD(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD(V18, V8, V8)
	TRIPLE_MOD(V9, V9)
	SUB_MOD(V18, V9, V9)
	QUAD_MOD(V10, V10)
	SUB_MOD(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD(V18, V17, V17)
	MOVD   576(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   576(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   576(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   576(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   576(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   576(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   576(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   576(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   576(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   576(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   576(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   576(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   576(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   576(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   576(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   576(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   600(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   600(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   600(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   600(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   600(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   600(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   600(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   600(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   600(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   600(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   600(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   600(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   600(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   600(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   600(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   600(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   624(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   624(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   624(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   624(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   624(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   624(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   624(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   624(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   624(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   624(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   624(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   624(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   624(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   624(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   624(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   624(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   RSP, R13
	VLD1.P 16(R13), [V18.S4]
	VLD1.P 16(R13), [V19.S4]
	VLD1.P 16(R13), [V20.S4]
	VLD1.P 16(R13), [V21.S4]
	VLD1.P 16(R13), [V22.S4]
	VLD1.P 16(R13), [V23.S4]
	VLD1.P 16(R13), [V24.S4]
	VLD1.P 16(R13), [V25.S4]
	ADD_MOD(V10, V18, V2)
	ADD_MOD(V11, V19, V3)
	ADD_MOD(V12, V20, V4)
	ADD_MOD(V13, V21, V5)
	ADD_MOD(V14, V22, V6)
	ADD_MOD(V15, V23, V7)
	ADD_MOD(V16, V24, V8)
	ADD_MOD(V17, V25, V9)
	ADD    $1, R8, R8
	CMP    $0x40, R8
	BNE    step_loop
	LSL    $7, R7, R13
	ADD    R2, R13, R9
	ADD    $0x20, R9, R10
	ADD    $0x20, R10, R11
	ADD    $0x20, R11, R12
	VMOV   V2.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V2.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V2.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V2.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V3.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V3.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V3.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V3.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V4.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V4.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V4.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V4.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V5.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V5.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V5.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V5.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V6.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V6.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V6.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V6.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V7.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V7.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V7.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V7.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V8.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V8.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V8.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V8.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V9.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V9.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V9.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V9.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	ADD    $1, R7, R7
	CMP    $0x4, R7
	BNE    batch_loop
	RET

TEXT ·permutation16x16xN_columns_arm64(SB), $128-48
	MOVD matrix+0(FP), R0
	MOVD roundKeys+8(FP), R1
	MOVD result+32(FP), R2
	MOVD $0x7f000001, R3
	MOVD $0x7effffff, R4
	VDUP R3, V0.S4
	VDUP R4, V1.S4
	MOVD $1, R5
	VDUP R5, V28.S4
	MOVD nbSteps+40(FP), R14

#define ADD_MOD_0(in0, in1, in2) \
	VADD  in0.S4, in1.S4, V30.S4 \
	VSUB  V0.S4, V30.S4, V31.S4  \
	VUMIN V30.S4, V31.S4, in2.S4 \

#define SUB_MOD_0(in0, in1, in2) \
	VSUB  in1.S4, in0.S4, V30.S4 \
	VADD  V0.S4, V30.S4, V31.S4  \
	VUMIN V30.S4, V31.S4, in2.S4 \

#define DOUBLE_MOD_0(in0, in1) \
	VSHL  $1, in0.S4, V30.S4     \
	VSUB  V0.S4, V30.S4, V31.S4  \
	VUMIN V30.S4, V31.S4, in1.S4 \

#define TRIPLE_MOD_0(in0, in1) \
DOUBLE_MOD_0(in0, V30)   \
ADD_MOD_0(V30, in0, in1) \

#define QUAD_MOD_0(in0, in1) \
DOUBLE_MOD_0(in0, V30) \
DOUBLE_MOD_0(V30, in1) \

#define MAT_MUL_4_0(in0, in1, in2, in3) \
	ADD_MOD_0(in0, in1, V18) \
	ADD_MOD_0(in2, in3, V19) \
	ADD_MOD_0(V18, V19, V20) \
	ADD_MOD_0(V20, in1, V21) \
	ADD_MOD_0(V20, in3, V22) \
	DOUBLE_MOD_0(in0, in3)   \
	ADD_MOD_0(in3, V22, in3) \
	DOUBLE_MOD_0(in2, in1)   \
	ADD_MOD_0(in1, V21, in1) \
	ADD_MOD_0(V18, V21, in0) \
	ADD_MOD_0(V19, V22, in2) \

#define MAT_MUL_EXT_0(in0, in1, in2, in3, in4, in5, in6, in7, in8, in9, in10, in11, in12, in13, in14, in15) \
MAT_MUL_4_0(in0, in1, in2, in3)     \
MAT_MUL_4_0(in4, in5, in6, in7)     \
MAT_MUL_4_0(in8, in9, in10, in11)   \
MAT_MUL_4_0(in12, in13, in14, in15) \
ADD_MOD_0(in0, in4, V18)            \
ADD_MOD_0(V18, in8, V18)            \
ADD_MOD_0(V18, in12, V18)           \
ADD_MOD_0(in1, in5, V19)            \
ADD_MOD_0(V19, in9, V19)            \
ADD_MOD_0(V19, in13, V19)           \
ADD_MOD_0(in2, in6, V20)            \
ADD_MOD_0(V20, in10, V20)           \
ADD_MOD_0(V20, in14, V20)           \
ADD_MOD_0(in3, in7, V21)            \
ADD_MOD_0(V21, in11, V21)           \
ADD_MOD_0(V21, in15, V21)           \
ADD_MOD_0(in0, V18, in0)            \
ADD_MOD_0(in1, V19, in1)            \
ADD_MOD_0(in2, V20, in2)            \
ADD_MOD_0(in3, V21, in3)            \
ADD_MOD_0(in4, V18, in4)            \
ADD_MOD_0(in5, V19, in5)            \
ADD_MOD_0(in6, V20, in6)            \
ADD_MOD_0(in7, V21, in7)            \
ADD_MOD_0(in8, V18, in8)            \
ADD_MOD_0(in9, V19, in9)            \
ADD_MOD_0(in10, V20, in10)          \
ADD_MOD_0(in11, V21, in11)          \
ADD_MOD_0(in12, V18, in12)          \
ADD_MOD_0(in13, V19, in13)          \
ADD_MOD_0(in14, V20, in14)          \
ADD_MOD_0(in15, V21, in15)          \

	MOVD $0, R7

batch_loop:
	VEOR V2.B16, V2.B16, V2.B16
	VEOR V3.B16, V3.B16, V3.B16
	VEOR V4.B16, V4.B16, V4.B16
	VEOR V5.B16, V5.B16, V5.B16
	VEOR V6.B16, V6.B16, V6.B16
	VEOR V7.B16, V7.B16, V7.B16
	VEOR V8.B16, V8.B16, V8.B16
	VEOR V9.B16, V9.B16, V9.B16
	VEOR V10.B16, V10.B16, V10.B16
	VEOR V11.B16, V11.B16, V11.B16
	VEOR V12.B16, V12.B16, V12.B16
	VEOR V13.B16, V13.B16, V13.B16
	VEOR V14.B16, V14.B16, V14.B16
	VEOR V15.B16, V15.B16, V15.B16
	VEOR V16.B16, V16.B16, V16.B16
	VEOR V17.B16, V17.B16, V17.B16
	MOVD $0, R8
	LSL  $4, R7, R13
	ADD  R0, R13, R9

step_loop:
	VLD1.P 16(R9), [V18.S4]
	ADD    $0x30, R9, R9
	VLD1.P 16(R9), [V19.S4]
	ADD    $0x30, R9, R9
	VLD1.P 16(R9), [V20.S4]
	ADD    $0x30, R9, R9
	VLD1.P 16(R9), [V21.S4]
	ADD    $0x30, R9, R9
	VLD1.P 16(R9), [V22.S4]
	ADD    $0x30, R9, R9
	VLD1.P 16(R9), [V23.S4]
	ADD    $0x30, R9, R9
	VLD1.P 16(R9), [V24.S4]
	ADD    $0x30, R9, R9
	VLD1.P 16(R9), [V25.S4]
	ADD    $0x30, R9, R9
	VMOV   V18.B16, V10.B16
	VMOV   V19.B16, V11.B16
	VMOV   V20.B16, V12.B16
	VMOV   V21.B16, V13.B16
	VMOV   V22.B16, V14.B16
	VMOV   V23.B16, V15.B16
	VMOV   V24.B16, V16.B16
	VMOV   V25.B16, V17.B16
	MOVD   RSP, R13
	VST1.P [V18.S4], 16(R13)
	VST1.P [V19.S4], 16(R13)
	VST1.P [V20.S4], 16(R13)
	VST1.P [V21.S4], 16(R13)
	VST1.P [V22.S4], 16(R13)
	VST1.P [V23.S4], 16(R13)
	VST1.P [V24.S4], 16(R13)
	VST1.P [V25.S4], 16(R13)
	MAT_MUL_EXT_0(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   0(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   0(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   0(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   0(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   0(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   0(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   0(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   0(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   0(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   0(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   0(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   0(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   0(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   0(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   0(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   0(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT_0(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   24(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   24(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   24(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   24(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   24(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   24(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   24(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   24(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   24(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   24(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   24(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   24(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   24(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   24(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   24(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   24(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT_0(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   48(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   48(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   48(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   48(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   48(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   48(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   48(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   48(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   48(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   48(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   48(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   48(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   48(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   48(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   48(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   48(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT_0(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   72(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   96(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   120(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   144(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   168(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   192(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   216(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   240(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   264(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   288(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   312(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   336(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   360(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   384(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   408(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   432(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   456(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   480(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   504(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   528(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   552(R1), R6
	VLD1R  (R6), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	ADD_MOD_0(V2, V3, V18)
	ADD_MOD_0(V4, V5, V19)
	ADD_MOD_0(V6, V7, V20)
	ADD_MOD_0(V8, V9, V21)
	ADD_MOD_0(V18, V19, V18)
	ADD_MOD_0(V20, V21, V20)
	ADD_MOD_0(V18, V20, V18)
	ADD_MOD_0(V10, V11, V22)
	ADD_MOD_0(V12, V13, V23)
	ADD_MOD_0(V14, V15, V24)
	ADD_MOD_0(V16, V17, V25)
	ADD_MOD_0(V22, V23, V22)
	ADD_MOD_0(V24, V25, V24)
	ADD_MOD_0(V22, V24, V22)
	ADD_MOD_0(V18, V22, V18)
	DOUBLE_MOD_0(V2, V2)
	SUB_MOD_0(V18, V2, V2)
	ADD_MOD_0(V18, V3, V3)
	DOUBLE_MOD_0(V4, V4)
	ADD_MOD_0(V18, V4, V4)
	VAND   V5.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf04a5               // UHADD V5.4S, V5.4S, V31.4S
	ADD_MOD_0(V18, V5, V5)
	TRIPLE_MOD_0(V6, V6)
	ADD_MOD_0(V18, V6, V6)
	QUAD_MOD_0(V7, V7)
	ADD_MOD_0(V18, V7, V7)
	VAND   V8.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0508               // UHADD V8.4S, V8.4S, V31.4S
	SUB_MOD_0(V18, V8, V8)
	TRIPLE_MOD_0(V9, V9)
	SUB_MOD_0(V18, V9, V9)
	QUAD_MOD_0(V10, V10)
	SUB_MOD_0(V18, V10, V10)
	WORD   $0x2f38a57e               // USHLL V30.2D, V11.2S, #24
	WORD   $0x6f38a57f               // USHLL2 V31.2D, V11.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	ADD_MOD_0(V18, V11, V11)
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	VAND   V12.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf058c               // UHADD V12.4S, V12.4S, V31.4S
	ADD_MOD_0(V18, V12, V12)
	WORD   $0x2f28a5be               // USHLL V30.2D, V13.2S, #8
	WORD   $0x6f28a5bf               // USHLL2 V31.2D, V13.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	ADD_MOD_0(V18, V13, V13)
	WORD   $0x2f38a5de               // USHLL V30.2D, V14.2S, #24
	WORD   $0x6f38a5df               // USHLL2 V31.2D, V14.4S, #24
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	SUB_MOD_0(V18, V14, V14)
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	VAND   V15.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf05ef               // UHADD V15.4S, V15.4S, V31.4S
	SUB_MOD_0(V18, V15, V15)
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	VAND   V16.B16, V28.B16, V30.B16
	VSHL   $31, V30.S4, V30.S4
	WORD   $0x4f2107de               // SSHR V30.4S, V30.4S, #31
	VAND   V0.B16, V30.B16, V31.B16
	WORD   $0x6ebf0610               // UHADD V16.4S, V16.4S, V31.4S
	SUB_MOD_0(V18, V16, V16)
	WORD   $0x2f28a63e               // USHLL V30.2D, V17.2S, #8
	WORD   $0x6f28a63f               // USHLL2 V31.2D, V17.4S, #8
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	SUB_MOD_0(V18, V17, V17)
	MOVD   576(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   576(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   576(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   576(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   576(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   576(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   576(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   576(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   576(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   576(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   576(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   576(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   576(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   576(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   576(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   576(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT_0(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   600(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   600(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   600(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   600(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   600(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   600(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   600(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   600(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   600(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   600(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   600(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   600(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   600(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   600(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   600(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   600(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT_0(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   624(R1), R6
	ADD    $0, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V2, V30, V2)
	WORD   $0x2ea2c05e               // UMULL V30.2D, V2.2S, V2.2S
	WORD   $0x6ea2c05f               // UMULL2 V31.2D, V2.4S, V2.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c05e               // UMULL V30.2D, V2.2S, V18.2S
	WORD   $0x6eb2c05f               // UMULL2 V31.2D, V2.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc2               // UZP2 V2.4S, V30.4S, V31.4S
	VSUB   V0.S4, V2.S4, V29.S4
	VUMIN  V2.S4, V29.S4, V2.S4
	MOVD   624(R1), R6
	ADD    $0x4, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V3, V30, V3)
	WORD   $0x2ea3c07e               // UMULL V30.2D, V3.2S, V3.2S
	WORD   $0x6ea3c07f               // UMULL2 V31.2D, V3.4S, V3.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c07e               // UMULL V30.2D, V3.2S, V18.2S
	WORD   $0x6eb2c07f               // UMULL2 V31.2D, V3.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc3               // UZP2 V3.4S, V30.4S, V31.4S
	VSUB   V0.S4, V3.S4, V29.S4
	VUMIN  V3.S4, V29.S4, V3.S4
	MOVD   624(R1), R6
	ADD    $0x8, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V4, V30, V4)
	WORD   $0x2ea4c09e               // UMULL V30.2D, V4.2S, V4.2S
	WORD   $0x6ea4c09f               // UMULL2 V31.2D, V4.4S, V4.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c09e               // UMULL V30.2D, V4.2S, V18.2S
	WORD   $0x6eb2c09f               // UMULL2 V31.2D, V4.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc4               // UZP2 V4.4S, V30.4S, V31.4S
	VSUB   V0.S4, V4.S4, V29.S4
	VUMIN  V4.S4, V29.S4, V4.S4
	MOVD   624(R1), R6
	ADD    $0xc, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V5, V30, V5)
	WORD   $0x2ea5c0be               // UMULL V30.2D, V5.2S, V5.2S
	WORD   $0x6ea5c0bf               // UMULL2 V31.2D, V5.4S, V5.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0be               // UMULL V30.2D, V5.2S, V18.2S
	WORD   $0x6eb2c0bf               // UMULL2 V31.2D, V5.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc5               // UZP2 V5.4S, V30.4S, V31.4S
	VSUB   V0.S4, V5.S4, V29.S4
	VUMIN  V5.S4, V29.S4, V5.S4
	MOVD   624(R1), R6
	ADD    $0x10, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V6, V30, V6)
	WORD   $0x2ea6c0de               // UMULL V30.2D, V6.2S, V6.2S
	WORD   $0x6ea6c0df               // UMULL2 V31.2D, V6.4S, V6.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0de               // UMULL V30.2D, V6.2S, V18.2S
	WORD   $0x6eb2c0df               // UMULL2 V31.2D, V6.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc6               // UZP2 V6.4S, V30.4S, V31.4S
	VSUB   V0.S4, V6.S4, V29.S4
	VUMIN  V6.S4, V29.S4, V6.S4
	MOVD   624(R1), R6
	ADD    $0x14, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V7, V30, V7)
	WORD   $0x2ea7c0fe               // UMULL V30.2D, V7.2S, V7.2S
	WORD   $0x6ea7c0ff               // UMULL2 V31.2D, V7.4S, V7.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c0fe               // UMULL V30.2D, V7.2S, V18.2S
	WORD   $0x6eb2c0ff               // UMULL2 V31.2D, V7.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc7               // UZP2 V7.4S, V30.4S, V31.4S
	VSUB   V0.S4, V7.S4, V29.S4
	VUMIN  V7.S4, V29.S4, V7.S4
	MOVD   624(R1), R6
	ADD    $0x18, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V8, V30, V8)
	WORD   $0x2ea8c11e               // UMULL V30.2D, V8.2S, V8.2S
	WORD   $0x6ea8c11f               // UMULL2 V31.2D, V8.4S, V8.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c11e               // UMULL V30.2D, V8.2S, V18.2S
	WORD   $0x6eb2c11f               // UMULL2 V31.2D, V8.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc8               // UZP2 V8.4S, V30.4S, V31.4S
	VSUB   V0.S4, V8.S4, V29.S4
	VUMIN  V8.S4, V29.S4, V8.S4
	MOVD   624(R1), R6
	ADD    $0x1c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V9, V30, V9)
	WORD   $0x2ea9c13e               // UMULL V30.2D, V9.2S, V9.2S
	WORD   $0x6ea9c13f               // UMULL2 V31.2D, V9.4S, V9.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c13e               // UMULL V30.2D, V9.2S, V18.2S
	WORD   $0x6eb2c13f               // UMULL2 V31.2D, V9.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bc9               // UZP2 V9.4S, V30.4S, V31.4S
	VSUB   V0.S4, V9.S4, V29.S4
	VUMIN  V9.S4, V29.S4, V9.S4
	MOVD   624(R1), R6
	ADD    $0x20, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V10, V30, V10)
	WORD   $0x2eaac15e               // UMULL V30.2D, V10.2S, V10.2S
	WORD   $0x6eaac15f               // UMULL2 V31.2D, V10.4S, V10.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c15e               // UMULL V30.2D, V10.2S, V18.2S
	WORD   $0x6eb2c15f               // UMULL2 V31.2D, V10.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bca               // UZP2 V10.4S, V30.4S, V31.4S
	VSUB   V0.S4, V10.S4, V29.S4
	VUMIN  V10.S4, V29.S4, V10.S4
	MOVD   624(R1), R6
	ADD    $0x24, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V11, V30, V11)
	WORD   $0x2eabc17e               // UMULL V30.2D, V11.2S, V11.2S
	WORD   $0x6eabc17f               // UMULL2 V31.2D, V11.4S, V11.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c17e               // UMULL V30.2D, V11.2S, V18.2S
	WORD   $0x6eb2c17f               // UMULL2 V31.2D, V11.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcb               // UZP2 V11.4S, V30.4S, V31.4S
	VSUB   V0.S4, V11.S4, V29.S4
	VUMIN  V11.S4, V29.S4, V11.S4
	MOVD   624(R1), R6
	ADD    $0x28, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V12, V30, V12)
	WORD   $0x2eacc19e               // UMULL V30.2D, V12.2S, V12.2S
	WORD   $0x6eacc19f               // UMULL2 V31.2D, V12.4S, V12.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c19e               // UMULL V30.2D, V12.2S, V18.2S
	WORD   $0x6eb2c19f               // UMULL2 V31.2D, V12.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcc               // UZP2 V12.4S, V30.4S, V31.4S
	VSUB   V0.S4, V12.S4, V29.S4
	VUMIN  V12.S4, V29.S4, V12.S4
	MOVD   624(R1), R6
	ADD    $0x2c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V13, V30, V13)
	WORD   $0x2eadc1be               // UMULL V30.2D, V13.2S, V13.2S
	WORD   $0x6eadc1bf               // UMULL2 V31.2D, V13.4S, V13.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1be               // UMULL V30.2D, V13.2S, V18.2S
	WORD   $0x6eb2c1bf               // UMULL2 V31.2D, V13.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcd               // UZP2 V13.4S, V30.4S, V31.4S
	VSUB   V0.S4, V13.S4, V29.S4
	VUMIN  V13.S4, V29.S4, V13.S4
	MOVD   624(R1), R6
	ADD    $0x30, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V14, V30, V14)
	WORD   $0x2eaec1de               // UMULL V30.2D, V14.2S, V14.2S
	WORD   $0x6eaec1df               // UMULL2 V31.2D, V14.4S, V14.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1de               // UMULL V30.2D, V14.2S, V18.2S
	WORD   $0x6eb2c1df               // UMULL2 V31.2D, V14.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bce               // UZP2 V14.4S, V30.4S, V31.4S
	VSUB   V0.S4, V14.S4, V29.S4
	VUMIN  V14.S4, V29.S4, V14.S4
	MOVD   624(R1), R6
	ADD    $0x34, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V15, V30, V15)
	WORD   $0x2eafc1fe               // UMULL V30.2D, V15.2S, V15.2S
	WORD   $0x6eafc1ff               // UMULL2 V31.2D, V15.4S, V15.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c1fe               // UMULL V30.2D, V15.2S, V18.2S
	WORD   $0x6eb2c1ff               // UMULL2 V31.2D, V15.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bcf               // UZP2 V15.4S, V30.4S, V31.4S
	VSUB   V0.S4, V15.S4, V29.S4
	VUMIN  V15.S4, V29.S4, V15.S4
	MOVD   624(R1), R6
	ADD    $0x38, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V16, V30, V16)
	WORD   $0x2eb0c21e               // UMULL V30.2D, V16.2S, V16.2S
	WORD   $0x6eb0c21f               // UMULL2 V31.2D, V16.4S, V16.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c21e               // UMULL V30.2D, V16.2S, V18.2S
	WORD   $0x6eb2c21f               // UMULL2 V31.2D, V16.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd0               // UZP2 V16.4S, V30.4S, V31.4S
	VSUB   V0.S4, V16.S4, V29.S4
	VUMIN  V16.S4, V29.S4, V16.S4
	MOVD   624(R1), R6
	ADD    $0x3c, R6, R13
	VLD1R  (R13), [V30.S4]
	ADD_MOD_0(V17, V30, V17)
	WORD   $0x2eb1c23e               // UMULL V30.2D, V17.2S, V17.2S
	WORD   $0x6eb1c23f               // UMULL2 V31.2D, V17.4S, V17.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd2               // UZP2 V18.4S, V30.4S, V31.4S
	VSUB   V0.S4, V18.S4, V29.S4
	VUMIN  V18.S4, V29.S4, V18.S4
	WORD   $0x2eb2c23e               // UMULL V30.2D, V17.2S, V18.2S
	WORD   $0x6eb2c23f               // UMULL2 V31.2D, V17.4S, V18.4S
	WORD   $0x4e9f1bdd               // UZP1 V29.4S, V30.4S, V31.4S
	WORD   $0x4ea19fbd               // MUL V29.4S, V29.4S, V1.4S
	WORD   $0x2ea0c3ba               // UMULL V26.2D, V29.2S, V0.2S
	WORD   $0x6ea0c3bb               // UMULL2 V27.2D, V29.4S, V0.4S
	VADD   V30.D2, V26.D2, V30.D2
	VADD   V31.D2, V27.D2, V31.D2
	WORD   $0x4e9f5bd1               // UZP2 V17.4S, V30.4S, V31.4S
	VSUB   V0.S4, V17.S4, V29.S4
	VUMIN  V17.S4, V29.S4, V17.S4
	MAT_MUL_EXT_0(V2, V3, V4, V5, V6, V7, V8, V9, V10, V11, V12, V13, V14, V15, V16, V17)
	MOVD   RSP, R13
	VLD1.P 16(R13), [V18.S4]
	VLD1.P 16(R13), [V19.S4]
	VLD1.P 16(R13), [V20.S4]
	VLD1.P 16(R13), [V21.S4]
	VLD1.P 16(R13), [V22.S4]
	VLD1.P 16(R13), [V23.S4]
	VLD1.P 16(R13), [V24.S4]
	VLD1.P 16(R13), [V25.S4]
	ADD_MOD_0(V10, V18, V2)
	ADD_MOD_0(V11, V19, V3)
	ADD_MOD_0(V12, V20, V4)
	ADD_MOD_0(V13, V21, V5)
	ADD_MOD_0(V14, V22, V6)
	ADD_MOD_0(V15, V23, V7)
	ADD_MOD_0(V16, V24, V8)
	ADD_MOD_0(V17, V25, V9)
	ADD    $1, R8, R8
	CMP    R14, R8
	BNE    step_loop
	LSL    $7, R7, R13
	ADD    R2, R13, R9
	ADD    $0x20, R9, R10
	ADD    $0x20, R10, R11
	ADD    $0x20, R11, R12
	VMOV   V2.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V2.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V2.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V2.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V3.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V3.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V3.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V3.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V4.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V4.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V4.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V4.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V5.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V5.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V5.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V5.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V6.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V6.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V6.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V6.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V7.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V7.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V7.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V7.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V8.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V8.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V8.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V8.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	VMOV   V9.S[0], R13
	MOVWU  R13, (R9)
	VMOV   V9.S[1], R13
	MOVWU  R13, (R10)
	VMOV   V9.S[2], R13
	MOVWU  R13, (R11)
	VMOV   V9.S[3], R13
	MOVWU  R13, (R12)
	ADD    $0x4, R9, R9
	ADD    $0x4, R10, R10
	ADD    $0x4, R11, R11
	ADD    $0x4, R12, R12
	ADD    $1, R7, R7
	CMP    $0x4, R7
	BNE    batch_loop
	RET
