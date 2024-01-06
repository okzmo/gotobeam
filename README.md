## Programming language to transform simple instruction to beam bytecode.

The idea to do this comes from [Tsoding](https://twitter.com/tsoding) just wanted to see if I could do it in go somehow.

To test it you simply need erlang and golang on your machine.

Write some simple instructions in a .iris file like "test.iris", an instruction look like this:

hello = 10

hello will be the name of the function and 10 is the return value.

When this is done you can just run the go file to transform this into beam bytecode and load the new "iris.beam" file into erlang. To do that simply do "erl" in your terminal then code:load_file(iris) and type: "iris:name_of_your_function()." this willreturn the value you put after the equal, in the above example it'll be 10.

That's all, maybe in the future I'll try to expand it further but as of right now it's the first time I wrote a compiler and I'm quite happy with it.
