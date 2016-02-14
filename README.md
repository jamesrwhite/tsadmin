tsadmin
========

Get a realtime overview of your MySQL servers. Very much a work in progress.

![preview](https://s3.amazonaws.com/f.cl.ly/items/2D3i151N3O0A1V3Y1d3P/Image%202016-02-14%20at%202.38.04%20pm.png?v=01b3f766)

Usage
------

```
PORT=8080 CONFIG_FILE=config/config.json go run tsadmin.go
```

Why 'tsadmin'
--------------

I'm not exactly sure where this name came from originally. There is an internal tool
with the same name as this that we use at [Miniclip](http://www.miniclip.com/) that
is probably almost as old as the company itself. It's very detailed and gets used a lot
but the code behind it is pretty scary, hence this.

TODO:
-----

- [x] Reads/Writes per second
- [ ] Replication lag
- [x] More detailed connection stats (per second & aborts)
- [ ] Sortable columns
- [ ] Groups/clusters
- [ ] Master detection
- [ ] Improved error handling
- [ ] Improved UI

License
--------

The MIT License (MIT)

Copyright (c) 2015 James White

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
