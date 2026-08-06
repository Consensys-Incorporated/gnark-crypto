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

#define MAT_MUL_EXT() \
	MAT_MUL_4(V2, V3, V4, V5)     \
	MAT_MUL_4(V6, V7, V8, V9)     \
	MAT_MUL_4(V10, V11, V12, V13) \
	MAT_MUL_4(V14, V15, V16, V17) \
	ADD_MOD(V2, V6, V18)          \
	ADD_MOD(V18, V10, V18)        \
	ADD_MOD(V18, V14, V18)        \
	ADD_MOD(V3, V7, V19)          \
	ADD_MOD(V19, V11, V19)        \
	ADD_MOD(V19, V15, V19)        \
	ADD_MOD(V4, V8, V20)          \
	ADD_MOD(V20, V12, V20)        \
	ADD_MOD(V20, V16, V20)        \
	ADD_MOD(V5, V9, V21)          \
	ADD_MOD(V21, V13, V21)        \
	ADD_MOD(V21, V17, V21)        \
	ADD_MOD(V2, V18, V2)          \
	ADD_MOD(V3, V19, V3)          \
	ADD_MOD(V4, V20, V4)          \
	ADD_MOD(V5, V21, V5)          \
	ADD_MOD(V6, V18, V6)          \
	ADD_MOD(V7, V19, V7)          \
	ADD_MOD(V8, V20, V8)          \
	ADD_MOD(V9, V21, V9)          \
	ADD_MOD(V10, V18, V10)        \
	ADD_MOD(V11, V19, V11)        \
	ADD_MOD(V12, V20, V12)        \
	ADD_MOD(V13, V21, V13)        \
	ADD_MOD(V14, V18, V14)        \
	ADD_MOD(V15, V19, V15)        \
	ADD_MOD(V16, V20, V16)        \
	ADD_MOD(V17, V21, V17)        \

#define MAT_MUL_INTERNAL() \
	ADD_MOD(V2, V3, V18)            \
	ADD_MOD(V4, V5, V19)            \
	ADD_MOD(V6, V7, V20)            \
	ADD_MOD(V8, V9, V21)            \
	ADD_MOD(V18, V19, V18)          \
	ADD_MOD(V20, V21, V20)          \
	ADD_MOD(V18, V20, V18)          \
	ADD_MOD(V10, V11, V22)          \
	ADD_MOD(V12, V13, V23)          \
	ADD_MOD(V14, V15, V24)          \
	ADD_MOD(V16, V17, V25)          \
	ADD_MOD(V22, V23, V22)          \
	ADD_MOD(V24, V25, V24)          \
	ADD_MOD(V22, V24, V22)          \
	ADD_MOD(V18, V22, V18)          \
	DOUBLE_MOD(V2, V2)              \
	SUB_MOD(V18, V2, V2)            \
	ADD_MOD(V18, V3, V3)            \
	DOUBLE_MOD(V4, V4)              \
	ADD_MOD(V18, V4, V4)            \
	VAND  V5.B16, V28.B16, V30.B16  \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf04a5               \
	ADD_MOD(V18, V5, V5)            \
	TRIPLE_MOD(V6, V6)              \
	ADD_MOD(V18, V6, V6)            \
	QUAD_MOD(V7, V7)                \
	ADD_MOD(V18, V7, V7)            \
	VAND  V8.B16, V28.B16, V30.B16  \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf0508               \
	SUB_MOD(V18, V8, V8)            \
	TRIPLE_MOD(V9, V9)              \
	SUB_MOD(V18, V9, V9)            \
	QUAD_MOD(V10, V10)              \
	SUB_MOD(V18, V10, V10)          \
	WORD  $0x2f38a57e               \
	WORD  $0x6f38a57f               \
	WORD  $0x4e9f1bdd               \
	WORD  $0x4ea19fbd               \
	WORD  $0x2ea0c3ba               \
	WORD  $0x6ea0c3bb               \
	VADD  V30.D2, V26.D2, V30.D2    \
	VADD  V31.D2, V27.D2, V31.D2    \
	WORD  $0x4e9f5bcb               \
	VSUB  V0.S4, V11.S4, V29.S4     \
	VUMIN V11.S4, V29.S4, V11.S4    \
	ADD_MOD(V18, V11, V11)          \
	VAND  V12.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf058c               \
	VAND  V12.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf058c               \
	VAND  V12.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf058c               \
	ADD_MOD(V18, V12, V12)          \
	WORD  $0x2f28a5be               \
	WORD  $0x6f28a5bf               \
	WORD  $0x4e9f1bdd               \
	WORD  $0x4ea19fbd               \
	WORD  $0x2ea0c3ba               \
	WORD  $0x6ea0c3bb               \
	VADD  V30.D2, V26.D2, V30.D2    \
	VADD  V31.D2, V27.D2, V31.D2    \
	WORD  $0x4e9f5bcd               \
	VSUB  V0.S4, V13.S4, V29.S4     \
	VUMIN V13.S4, V29.S4, V13.S4    \
	ADD_MOD(V18, V13, V13)          \
	WORD  $0x2f38a5de               \
	WORD  $0x6f38a5df               \
	WORD  $0x4e9f1bdd               \
	WORD  $0x4ea19fbd               \
	WORD  $0x2ea0c3ba               \
	WORD  $0x6ea0c3bb               \
	VADD  V30.D2, V26.D2, V30.D2    \
	VADD  V31.D2, V27.D2, V31.D2    \
	WORD  $0x4e9f5bce               \
	VSUB  V0.S4, V14.S4, V29.S4     \
	VUMIN V14.S4, V29.S4, V14.S4    \
	SUB_MOD(V18, V14, V14)          \
	VAND  V15.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf05ef               \
	VAND  V15.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf05ef               \
	VAND  V15.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf05ef               \
	SUB_MOD(V18, V15, V15)          \
	VAND  V16.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf0610               \
	VAND  V16.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf0610               \
	VAND  V16.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf0610               \
	VAND  V16.B16, V28.B16, V30.B16 \
	VSHL  $31, V30.S4, V30.S4       \
	WORD  $0x4f2107de               \
	VAND  V0.B16, V30.B16, V31.B16  \
	WORD  $0x6ebf0610               \
	SUB_MOD(V18, V16, V16)          \
	WORD  $0x2f28a63e               \
	WORD  $0x6f28a63f               \
	WORD  $0x4e9f1bdd               \
	WORD  $0x4ea19fbd               \
	WORD  $0x2ea0c3ba               \
	WORD  $0x6ea0c3bb               \
	VADD  V30.D2, V26.D2, V30.D2    \
	VADD  V31.D2, V27.D2, V31.D2    \
	WORD  $0x4e9f5bd1               \
	VSUB  V0.S4, V17.S4, V29.S4     \
	VUMIN V17.S4, V29.S4, V17.S4    \
	SUB_MOD(V18, V17, V17)          \

#define FULL_ROUND(in0) \
	MOVD  in0(R1), R6            \
	VLD1R (R6), [V30.S4]         \
	ADD_MOD(V2, V30, V2)         \
	WORD  $0x2ea2c05e            \
	WORD  $0x6ea2c05f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c05e            \
	WORD  $0x6eb2c05f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc2            \
	VSUB  V0.S4, V2.S4, V29.S4   \
	VUMIN V2.S4, V29.S4, V2.S4   \
	ADD   $0x4, R6, R13          \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V3, V30, V3)         \
	WORD  $0x2ea3c07e            \
	WORD  $0x6ea3c07f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c07e            \
	WORD  $0x6eb2c07f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc3            \
	VSUB  V0.S4, V3.S4, V29.S4   \
	VUMIN V3.S4, V29.S4, V3.S4   \
	ADD   $0x8, R6, R13          \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V4, V30, V4)         \
	WORD  $0x2ea4c09e            \
	WORD  $0x6ea4c09f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c09e            \
	WORD  $0x6eb2c09f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc4            \
	VSUB  V0.S4, V4.S4, V29.S4   \
	VUMIN V4.S4, V29.S4, V4.S4   \
	ADD   $0xc, R6, R13          \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V5, V30, V5)         \
	WORD  $0x2ea5c0be            \
	WORD  $0x6ea5c0bf            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c0be            \
	WORD  $0x6eb2c0bf            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc5            \
	VSUB  V0.S4, V5.S4, V29.S4   \
	VUMIN V5.S4, V29.S4, V5.S4   \
	ADD   $0x10, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V6, V30, V6)         \
	WORD  $0x2ea6c0de            \
	WORD  $0x6ea6c0df            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c0de            \
	WORD  $0x6eb2c0df            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc6            \
	VSUB  V0.S4, V6.S4, V29.S4   \
	VUMIN V6.S4, V29.S4, V6.S4   \
	ADD   $0x14, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V7, V30, V7)         \
	WORD  $0x2ea7c0fe            \
	WORD  $0x6ea7c0ff            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c0fe            \
	WORD  $0x6eb2c0ff            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc7            \
	VSUB  V0.S4, V7.S4, V29.S4   \
	VUMIN V7.S4, V29.S4, V7.S4   \
	ADD   $0x18, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V8, V30, V8)         \
	WORD  $0x2ea8c11e            \
	WORD  $0x6ea8c11f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c11e            \
	WORD  $0x6eb2c11f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc8            \
	VSUB  V0.S4, V8.S4, V29.S4   \
	VUMIN V8.S4, V29.S4, V8.S4   \
	ADD   $0x1c, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V9, V30, V9)         \
	WORD  $0x2ea9c13e            \
	WORD  $0x6ea9c13f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c13e            \
	WORD  $0x6eb2c13f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc9            \
	VSUB  V0.S4, V9.S4, V29.S4   \
	VUMIN V9.S4, V29.S4, V9.S4   \
	ADD   $0x20, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V10, V30, V10)       \
	WORD  $0x2eaac15e            \
	WORD  $0x6eaac15f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c15e            \
	WORD  $0x6eb2c15f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bca            \
	VSUB  V0.S4, V10.S4, V29.S4  \
	VUMIN V10.S4, V29.S4, V10.S4 \
	ADD   $0x24, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V11, V30, V11)       \
	WORD  $0x2eabc17e            \
	WORD  $0x6eabc17f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c17e            \
	WORD  $0x6eb2c17f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bcb            \
	VSUB  V0.S4, V11.S4, V29.S4  \
	VUMIN V11.S4, V29.S4, V11.S4 \
	ADD   $0x28, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V12, V30, V12)       \
	WORD  $0x2eacc19e            \
	WORD  $0x6eacc19f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c19e            \
	WORD  $0x6eb2c19f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bcc            \
	VSUB  V0.S4, V12.S4, V29.S4  \
	VUMIN V12.S4, V29.S4, V12.S4 \
	ADD   $0x2c, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V13, V30, V13)       \
	WORD  $0x2eadc1be            \
	WORD  $0x6eadc1bf            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c1be            \
	WORD  $0x6eb2c1bf            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bcd            \
	VSUB  V0.S4, V13.S4, V29.S4  \
	VUMIN V13.S4, V29.S4, V13.S4 \
	ADD   $0x30, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V14, V30, V14)       \
	WORD  $0x2eaec1de            \
	WORD  $0x6eaec1df            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c1de            \
	WORD  $0x6eb2c1df            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bce            \
	VSUB  V0.S4, V14.S4, V29.S4  \
	VUMIN V14.S4, V29.S4, V14.S4 \
	ADD   $0x34, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V15, V30, V15)       \
	WORD  $0x2eafc1fe            \
	WORD  $0x6eafc1ff            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c1fe            \
	WORD  $0x6eb2c1ff            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bcf            \
	VSUB  V0.S4, V15.S4, V29.S4  \
	VUMIN V15.S4, V29.S4, V15.S4 \
	ADD   $0x38, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V16, V30, V16)       \
	WORD  $0x2eb0c21e            \
	WORD  $0x6eb0c21f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c21e            \
	WORD  $0x6eb2c21f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd0            \
	VSUB  V0.S4, V16.S4, V29.S4  \
	VUMIN V16.S4, V29.S4, V16.S4 \
	ADD   $0x3c, R6, R13         \
	VLD1R (R13), [V30.S4]        \
	ADD_MOD(V17, V30, V17)       \
	WORD  $0x2eb1c23e            \
	WORD  $0x6eb1c23f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c23e            \
	WORD  $0x6eb2c23f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd1            \
	VSUB  V0.S4, V17.S4, V29.S4  \
	VUMIN V17.S4, V29.S4, V17.S4 \
	MAT_MUL_EXT()                \

#define PARTIAL_ROUND(in0) \
	MOVD  in0(R1), R6            \
	VLD1R (R6), [V30.S4]         \
	ADD_MOD(V2, V30, V2)         \
	WORD  $0x2ea2c05e            \
	WORD  $0x6ea2c05f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bd2            \
	VSUB  V0.S4, V18.S4, V29.S4  \
	VUMIN V18.S4, V29.S4, V18.S4 \
	WORD  $0x2eb2c05e            \
	WORD  $0x6eb2c05f            \
	WORD  $0x4e9f1bdd            \
	WORD  $0x4ea19fbd            \
	WORD  $0x2ea0c3ba            \
	WORD  $0x6ea0c3bb            \
	VADD  V30.D2, V26.D2, V30.D2 \
	VADD  V31.D2, V27.D2, V31.D2 \
	WORD  $0x4e9f5bc2            \
	VSUB  V0.S4, V2.S4, V29.S4   \
	VUMIN V2.S4, V29.S4, V2.S4   \
	MAT_MUL_INTERNAL()           \

#define ZERO_STATE() \
	VEOR V2.B16, V2.B16, V2.B16    \
	VEOR V3.B16, V3.B16, V3.B16    \
	VEOR V4.B16, V4.B16, V4.B16    \
	VEOR V5.B16, V5.B16, V5.B16    \
	VEOR V6.B16, V6.B16, V6.B16    \
	VEOR V7.B16, V7.B16, V7.B16    \
	VEOR V8.B16, V8.B16, V8.B16    \
	VEOR V9.B16, V9.B16, V9.B16    \
	VEOR V10.B16, V10.B16, V10.B16 \
	VEOR V11.B16, V11.B16, V11.B16 \
	VEOR V12.B16, V12.B16, V12.B16 \
	VEOR V13.B16, V13.B16, V13.B16 \
	VEOR V14.B16, V14.B16, V14.B16 \
	VEOR V15.B16, V15.B16, V15.B16 \
	VEOR V16.B16, V16.B16, V16.B16 \
	VEOR V17.B16, V17.B16, V17.B16 \

#define SAVE_INPUTS() \
	VMOV   V18.B16, V10.B16  \
	VMOV   V19.B16, V11.B16  \
	VMOV   V20.B16, V12.B16  \
	VMOV   V21.B16, V13.B16  \
	VMOV   V22.B16, V14.B16  \
	VMOV   V23.B16, V15.B16  \
	VMOV   V24.B16, V16.B16  \
	VMOV   V25.B16, V17.B16  \
	MOVD   RSP, R13          \
	VST1.P [V18.S4], 16(R13) \
	VST1.P [V19.S4], 16(R13) \
	VST1.P [V20.S4], 16(R13) \
	VST1.P [V21.S4], 16(R13) \
	VST1.P [V22.S4], 16(R13) \
	VST1.P [V23.S4], 16(R13) \
	VST1.P [V24.S4], 16(R13) \
	VST1.P [V25.S4], 16(R13) \

#define FEED_FORWARD() \
	MOVD   RSP, R13          \
	VLD1.P 16(R13), [V18.S4] \
	VLD1.P 16(R13), [V19.S4] \
	VLD1.P 16(R13), [V20.S4] \
	VLD1.P 16(R13), [V21.S4] \
	VLD1.P 16(R13), [V22.S4] \
	VLD1.P 16(R13), [V23.S4] \
	VLD1.P 16(R13), [V24.S4] \
	VLD1.P 16(R13), [V25.S4] \
	ADD_MOD(V10, V18, V2)    \
	ADD_MOD(V11, V19, V3)    \
	ADD_MOD(V12, V20, V4)    \
	ADD_MOD(V13, V21, V5)    \
	ADD_MOD(V14, V22, V6)    \
	ADD_MOD(V15, V23, V7)    \
	ADD_MOD(V16, V24, V8)    \
	ADD_MOD(V17, V25, V9)    \

#define STORE_4LANES(in0) \
	VMOV  in0.S[0], R13  \
	MOVWU R13, (R9)      \
	VMOV  in0.S[1], R13  \
	MOVWU R13, (R10)     \
	VMOV  in0.S[2], R13  \
	MOVWU R13, (R11)     \
	VMOV  in0.S[3], R13  \
	MOVWU R13, (R12)     \
	ADD   $0x4, R9, R9   \
	ADD   $0x4, R10, R10 \
	ADD   $0x4, R11, R11 \
	ADD   $0x4, R12, R12 \

#define LOAD_4LANES(in0) \
	MOVWU (R9), R13      \
	VMOV  R13, in0.S[0]  \
	MOVWU (R10), R13     \
	VMOV  R13, in0.S[1]  \
	MOVWU (R11), R13     \
	VMOV  R13, in0.S[2]  \
	MOVWU (R12), R13     \
	VMOV  R13, in0.S[3]  \
	ADD   $0x4, R9, R9   \
	ADD   $0x4, R10, R10 \
	ADD   $0x4, R11, R11 \
	ADD   $0x4, R12, R12 \

	MOVD $0, R7

batch_loop:
	ZERO_STATE()
	MOVD $0, R8
	LSL  $13, R7, R13
	ADD  R0, R13, R9
	ADD  $0x800, R9, R10
	ADD  $0x800, R10, R11
	ADD  $0x800, R11, R12

step_loop:
	LOAD_4LANES(V18)
	LOAD_4LANES(V19)
	LOAD_4LANES(V20)
	LOAD_4LANES(V21)
	LOAD_4LANES(V22)
	LOAD_4LANES(V23)
	LOAD_4LANES(V24)
	LOAD_4LANES(V25)
	SAVE_INPUTS()
	MAT_MUL_EXT()
	FULL_ROUND(0)
	FULL_ROUND(24)
	FULL_ROUND(48)
	PARTIAL_ROUND(72)
	PARTIAL_ROUND(96)
	PARTIAL_ROUND(120)
	PARTIAL_ROUND(144)
	PARTIAL_ROUND(168)
	PARTIAL_ROUND(192)
	PARTIAL_ROUND(216)
	PARTIAL_ROUND(240)
	PARTIAL_ROUND(264)
	PARTIAL_ROUND(288)
	PARTIAL_ROUND(312)
	PARTIAL_ROUND(336)
	PARTIAL_ROUND(360)
	PARTIAL_ROUND(384)
	PARTIAL_ROUND(408)
	PARTIAL_ROUND(432)
	PARTIAL_ROUND(456)
	PARTIAL_ROUND(480)
	PARTIAL_ROUND(504)
	PARTIAL_ROUND(528)
	PARTIAL_ROUND(552)
	FULL_ROUND(576)
	FULL_ROUND(600)
	FULL_ROUND(624)
	FEED_FORWARD()
	ADD $1, R8, R8
	CMP $0x40, R8
	BNE step_loop
	LSL $7, R7, R13
	ADD R2, R13, R9
	ADD $0x20, R9, R10
	ADD $0x20, R10, R11
	ADD $0x20, R11, R12
	STORE_4LANES(V2)
	STORE_4LANES(V3)
	STORE_4LANES(V4)
	STORE_4LANES(V5)
	STORE_4LANES(V6)
	STORE_4LANES(V7)
	STORE_4LANES(V8)
	STORE_4LANES(V9)
	ADD $1, R7, R7
	CMP $0x4, R7
	BNE batch_loop
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
	MOVD $0, R7

batch_loop:
	ZERO_STATE()
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
	SAVE_INPUTS()
	MAT_MUL_EXT()
	FULL_ROUND(0)
	FULL_ROUND(24)
	FULL_ROUND(48)
	PARTIAL_ROUND(72)
	PARTIAL_ROUND(96)
	PARTIAL_ROUND(120)
	PARTIAL_ROUND(144)
	PARTIAL_ROUND(168)
	PARTIAL_ROUND(192)
	PARTIAL_ROUND(216)
	PARTIAL_ROUND(240)
	PARTIAL_ROUND(264)
	PARTIAL_ROUND(288)
	PARTIAL_ROUND(312)
	PARTIAL_ROUND(336)
	PARTIAL_ROUND(360)
	PARTIAL_ROUND(384)
	PARTIAL_ROUND(408)
	PARTIAL_ROUND(432)
	PARTIAL_ROUND(456)
	PARTIAL_ROUND(480)
	PARTIAL_ROUND(504)
	PARTIAL_ROUND(528)
	PARTIAL_ROUND(552)
	FULL_ROUND(576)
	FULL_ROUND(600)
	FULL_ROUND(624)
	FEED_FORWARD()
	ADD    $1, R8, R8
	CMP    R14, R8
	BNE    step_loop
	LSL    $7, R7, R13
	ADD    R2, R13, R9
	ADD    $0x20, R9, R10
	ADD    $0x20, R10, R11
	ADD    $0x20, R11, R12
	STORE_4LANES(V2)
	STORE_4LANES(V3)
	STORE_4LANES(V4)
	STORE_4LANES(V5)
	STORE_4LANES(V6)
	STORE_4LANES(V7)
	STORE_4LANES(V8)
	STORE_4LANES(V9)
	ADD    $1, R7, R7
	CMP    $0x4, R7
	BNE    batch_loop
	RET
