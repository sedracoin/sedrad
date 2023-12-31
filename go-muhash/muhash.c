//
// Created by elichai2 on 2/15/21.
//

#undef NDEBUG
#include <assert.h>
#include <stdio.h>
#include "muhash.h"

#define MAX_PRIME_DIFF 1103717

/** Extract the lowest limb of [low,high,carry] into n, and left shift the number by 1 limb. */
static inline void extract3(limb_t *low, limb_t *high, limb_t *carry, limb_t *n) {
    *n = *low;
    *low = *high;
    *high = *carry;
    *carry = 0;
}

/** [low,high] = a * b */
static inline void mul(limb_t *low, limb_t *high, const limb_t *a, const limb_t *b) {
    double_limb_t t = (double_limb_t) *a * *b;
    *high = t >> LIMB_SIZE;
    *low = t;
}

/* [c0,c1,c2] += n * [d0,d1,d2]. c2 is 0 initially */
static inline void mulnadd3(limb_t *c0, limb_t *c1, limb_t *c2, limb_t *d0, limb_t *d1, limb_t *d2, const limb_t n) {
    double_limb_t t = (double_limb_t) *d0 * n + *c0;
    *c0 = t;
    t >>= LIMB_SIZE;

    t += (double_limb_t) *d1 * n + *c1;
    *c1 = t;
    t >>= LIMB_SIZE;
    *c2 = t + *d2 * n;
}

/* [low,high] *= n */
static inline void muln2(limb_t *low, limb_t *high, const limb_t n) {
    double_limb_t t = (double_limb_t) *low * n;
    *low = t;

    t >>= LIMB_SIZE;
    t += (double_limb_t) *high * n;
    *high = t;
}

/** [low,high,carry] += a * b */
static inline void muladd3(limb_t *low, limb_t *high, limb_t *carry, const limb_t *a, const limb_t *b) {
    double_limb_t t = (double_limb_t) *a * *b;
    limb_t th = t >> LIMB_SIZE;
    limb_t tl = t;

    *low += tl;
    th += (*low < tl) ? 1 : 0;
    *high += th;
    *carry += (*high < th) ? 1 : 0;
}

/** [low,high,carry] += 2 * a * b */
static inline void muldbladd3(limb_t *low, limb_t *high, limb_t *carry, const limb_t *a, const limb_t *b) {
    double_limb_t t = (double_limb_t) *a * *b;
    limb_t th = t >> LIMB_SIZE;
    limb_t tl = t;

    *low += tl;
    limb_t tt = th + ((*low < tl) ? 1 : 0);
    *high += tt;
    *carry += (*high < tt) ? 1 : 0;

    *low += tl;
    th += (*low < tl) ? 1 : 0;
    *high += th;
    *carry += (*high < th) ? 1 : 0;
}

/**
 * Add limb a to [low,high]: [low,high] += a. Then extract the lowest
 * limb of [low,high] into n, and left shift the number by 1 limb.
 * */
static inline void addnextract2(limb_t *low, limb_t *high, const limb_t *a, limb_t *n) {
    limb_t carry = 0;

// add
    *low += *a;
    if (*low < *a) {
        *high += 1;

// Handle case when high has overflown
        if (*high == 0)
            carry = 1;
    }

// extract
    *n = *low;
    *low = *high;
    *high = carry;
}
/** Indicates wether d is larger than the modulus. */
static inline int Num3072_IsOverflow(const Num3072 *this) {
    if (this->limbs[0] <= LIMB_MAX - MAX_PRIME_DIFF) return 0;
    for (int i = 1; i < LIMBS; ++i) {
        if (this->limbs[i] != LIMB_MAX) return 0;
    }
    return 1;
}

static inline void Num3072_SetToOne(Num3072 *this) {
    this->limbs[0] = 1;
    for (int i = 1; i < LIMBS; ++i) this->limbs[i] = 0;
}

void Num3072_FullReduce(Num3072 *this) {
    limb_t low = MAX_PRIME_DIFF;
    limb_t high = 0;
    for (int i = 0; i < LIMBS; ++i) {
        addnextract2(&low, &high, &this->limbs[i], &this->limbs[i]);
    }
}

void Num3072_Multiply(Num3072 *this, const Num3072 *a) {
    limb_t carryLow = 0, carryHigh = 0, carryHighest = 0;
    Num3072 tmp;

    /* Compute limbs 0..N-2 of this*a into tmp, including one reduction. */
    for (int j = 0; j < LIMBS - 1; ++j) {
        limb_t low = 0, high = 0, carry = 0;
        mul(&low, &high, &this->limbs[1 + j], &a->limbs[LIMBS + j - (1 + j)]);
        for (int i = 2 + j; i < LIMBS; ++i)
            muladd3(&low, &high, &carry, &this->limbs[i], &a->limbs[LIMBS + j - i]);

        mulnadd3(&carryLow, &carryHigh, &carryHighest, &low, &high, &carry, MAX_PRIME_DIFF);
        for (int i = 0; i < j + 1; ++i)
            muladd3(&carryLow, &carryHigh, &carryHighest, &this->limbs[i], &a->limbs[j - i]);

        extract3(&carryLow, &carryHigh, &carryHighest, &tmp.limbs[j]);
    }

    /* Compute limb N-1 of a*b into tmp. */
    assert(carryHighest == 0);
    for (int i = 0; i < LIMBS; ++i)
        muladd3(&carryLow, &carryHigh, &carryHighest, &this->limbs[i], &a->limbs[LIMBS - 1 - i]);
    extract3(&carryLow, &carryHigh, &carryHighest, &tmp.limbs[LIMBS - 1]);

    /* Perform a second reduction. */
    muln2(&carryLow, &carryHigh, MAX_PRIME_DIFF);
    for (int j = 0; j < LIMBS; ++j) {
        addnextract2(&carryLow, &carryHigh, &tmp.limbs[j], &this->limbs[j]);
    }

    assert(carryHigh == 0);
    assert(carryLow == 0 || carryLow == 1);

    /* Perform up to two more reductions if the internal state has already
     * overflown the MAX of Num3072 or if it is larger than the modulus or
     * if both are the case.
     * */
    if (Num3072_IsOverflow(this)) Num3072_FullReduce(this);
    if (carryLow) Num3072_FullReduce(this);
}
