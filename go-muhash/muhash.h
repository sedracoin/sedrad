//
// Created by elichai2 on 2/15/21.
//

#ifndef RUST_MUHASH_MUHASH_H
#define RUST_MUHASH_MUHASH_H
#include <stdint.h>


#if defined(__GNUC__)
#define GNUC_EXTENSION __extension__
#else
#define GNUC_EXTENSION
#endif

#if defined(UINT128_MAX) || defined(__SIZEOF_INT128__)
typedef uint64_t limb_t;
GNUC_EXTENSION typedef unsigned __int128 double_limb_t;
#define LIMB_SIZE 64
#define LIMBS 48
#define LIMB_MAX UINT64_MAX
#else
typedef uint32_t limb_t;
typedef uint64_t double_limb_t;
#define LIMB_SIZE 32
#define LIMBS 96
#define LIMB_MAX UINT32_MAX
#endif

typedef struct Num3072 {
    limb_t limbs[LIMBS];
} Num3072;

void Num3072_Multiply(Num3072* this, const Num3072* a);
void Num3072_Divide(Num3072* this, const Num3072* a);
Num3072 Num3072_GetInverse(const Num3072 *this);
void Num3072_FullReduce(Num3072* this);

#endif //RUST_MUHASH_MUHASH_H

