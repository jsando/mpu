# My todo list

- Defining a local with the same name as a global should be a warning
- Errors need to print the offending line, but text/scanner doesn't maintain it
- Bug in line # on errors, prints 1 line less than it should
- cleanup
    - lexer
        - text.scanner leaves quotes on strings, ticks on chars, etc.  It should have an IntValue(), CharValue(), StringValue(), etc.
        - would token category help?  directive, opcode, etc
        - don't use text.Scanner ... just use my own.  Use $ instead of 0x, ; instead of //.
    - parser
        - review all the parse functions and make sure they follow the same pattern ... do they call lexer.next?
        - sometimes I use tok := lexer.next and sometimes lexer.tok
- Really wish I had an indirect-indexed mode, when given a pointer to a struct I need constant offsets off the pointer
    - Suppose I could add a 1 byte (uint8) to relative-indirect, if none specified its zero?  My lovely byte savings go away :(
    - Tried it on some sample code and it cleans it up a lot
    - However, in writing some "real" code it was already pretty easy to hit the signed 8 bit offset limit for branches, this would make it even worse.
- Run from command line, ie pass argc, char **argv somehow?
- If there were a way to write unit tests in mpu, for mpu programs, that would provide an easy way to
  iteratively code and run them.
- monitor needs a way to view stack contents ... not sure how though unless we know whether they are bytes or words
- could I actually build a debugger that could inspect variables?

- Can I separate an abstraction for memory that takes all the byte/word/address abstractions into its own thing

I made a Memory abstraction and separated that and its tests from Machine.  It did clean some things up,
but the interface has too much in it.  Several implementations of memory have a "panic('not supported')", which
certainly seems like the smell of a over-complicated interface.  Need to refactor it into multiple smaller
interfaces but just now I don't see the right abstractions.

The pc/sp/fp use the Register abstraction to all support put/get byte/word consistently.  They use a &int so I don't have to 
give up performance of straight field access just to bump the pc.  Ie, reading/writing the registers via memory access should
be the exception not the norm.

The device io is also a memory but it doesn't support byte writes.  I do like how this cleaned up machine however, because now
all the io sits behind a single IOdispatcher.  In a physical machine there would be some sort of bus or dma controller as well.

I moved the monitor code into Monitor, which for now is in the root of mpu/.  

Machine is now very focused on just execution of its own opcodes.  Although maybe instead of machine, it should be MPU, or processor,
or some other name ... the machine includes the IO, memory, etc.  Machine here is just the execution engine.

