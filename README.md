This project is closed. If I begin again, it will be from scratch and using lessons learned while writing go.uik. Of course, if I begin again it will still be called go.uik.

#go.uik

A concurrent UI kit written in pure go.

This project is in its infancy. Feel free to experiment, but don't expect too much yet.

There is a [google group](https://groups.google.com/forum/?fromgroups#!forum/go-uik) dedicated to this project.

* * *

###A concurrent UI kit

Every component visible on the screen is backed by a *Block*. Every collection of components is backed by a *Foundation*. All Blocks are built upon a Foundation, but a Foundation itself is made up of a Block, which must itself be laid upon a Foundation. The only exception to this rule is the WindowFoundation.

All communication between Foundations, Blocks and the widgets and layouts composed of them is done via non-blocking channel communication.

While this is a break from the typical polymorphism approach, the result is that a component that stalls while processing input cannot get in the way of other components.

