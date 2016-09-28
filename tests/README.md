# pio/tests - smoke tests

This directory contains tests requirement hardware to interact with. That is,
each subdirectory is an executable that verify that both the driver and the
hardware work. The executable must exit with a return code of 0 when successful
and non-zero when an error is detected to enable an eventual future automated
testing lab.

Unit tests must be in the package defining the driver, not here.
