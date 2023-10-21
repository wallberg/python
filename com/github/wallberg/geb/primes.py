# -*- coding: utf-8 -*-

from dataclasses import dataclass
from typing import Any

from ordered_set import OrderedSet

''' Recursively enumerate primes, GEB Chapter 3.'''


@dataclass(frozen=True)
class P:
    x: int


@dataclass(frozen=True)
class DND:
    x: int
    y: int


@dataclass(frozen=True)
class DF:
    x: int
    y: int


theorems = OrderedSet()


def add(t, *s):
    ''' Indicate that t has been derived from s. '''

    n = theorems.add(t)

    # # Print to screen if this is the first derivation of t
    # if n == len(theorems)-1:
    #     if not s:
    #         print(t)
    #     else:
    #         print('{0} <-- {1}'.format(t, ' + '.join(str(x) for x in s)))


def axiom_schema():
    x = 2
    while True:
        for y in range(1, x):
            yield DND(x, y)
        x += 1


# Axioms
add(P(2))

asc = axiom_schema()

# Iterate over the queue
for n, s in enumerate(theorems, start=1):

    # Rules
    if isinstance(s, DND):

        # Rule: xDNDy --> xDNDxy
        add(DND(s.x, s.x+s.y), s)

        # Rule: --DNDz --> zDF--
        if s.x == 2:
            add(DF(s.y, 2), s)

        # Rule: zDFx and x-DNDz --> zDFx-
        if (u := DF(s.y, s.x-1)) in theorems:
            add(DF(s.y, s.x-1), u, s)

    elif isinstance(s, DF):

        # Rule: zDFx and x-DNDz --> zDFx-
        if (u := DND(s.y+1, s.x)) in theorems:
            add(DF(s.x, s.y+1), s, u)

        if s.x == s.y+1:
            p = P(s.x)
            add(p, s)
            print(p)

    # Add next Axiom from Schema
    add(next(asc))
