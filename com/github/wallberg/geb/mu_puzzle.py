# -*- coding: utf-8 -*-

import re

from ordered_set import OrderedSet

''' The MU-puzzle, GEB Chapter 1. '''

system = OrderedSet()

# Axioms
system.add('MI')


def add(s, t):
    ''' Indicate that t has been derived from s. '''

    n = system.add(t)

    # Print to screen if this is the first derivation of t
    if n == len(system)-1:
        print(f'{s} -> {t}')


# Iterate over each theorem in the system
for s in system:

    # Rule I: xI -> xIU
    if s.endswith('I'):
        add(s, s + 'U')

    # Rule II: Mx -> Mxx
    if s.startswith('M'):
        x = s[1:]
        add(s, 'M' + x + x)

    # Rule III: xIIIy -> xUy
    for m in re.finditer('III', s):
        add(s, s[:m.start()] + 'U' + s[m.end():])

    # Rule IV: xUUy -> xy
    for m in re.finditer('UU', s):
        add(s, s[:m.start()] + s[m.end():])

    if all(s in system for s in ['MUIIU']):
        break
