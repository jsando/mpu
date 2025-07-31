# VSCode Extension API Design for MPU Debugger

## Overview

This document outlines the design for a VSCode extension that integrates with the MPU debugger, providing a visual debugging experience for MPU assembly programs.

## Architecture

The VSCode extension will communicate with the MPU debugger through a Debug Adapter Protocol (DAP) server. This allows for a clean separation between the debugger backend and the VSCode frontend.

```
┌─────────────────┐     DAP      ┌──────────────────┐     ┌─────────────┐
│ VSCode Extension│◄────────────►│ MPU DAP Server   │◄───►│ MPU Debugger│
└─────────────────┘              └──────────────────┘     └─────────────┘
```

## DAP Server Implementation

Create a new package `debugger/dap` that implements the Debug Adapter Protocol:

```go
package dap

type DAPServer struct {
    debugger *debugger.Debugger
    // ... DAP protocol implementation
}

// Key methods to implement:
// - Initialize
// - Launch/Attach
// - SetBreakpoints
// - Continue/Next/StepIn/StepOut
// - StackTrace
// - Scopes
// - Variables
// - Evaluate
// - Disconnect
```

## VSCode Extension Features

### 1. Launch Configuration
```json
{
    "type": "mpu",
    "request": "launch",
    "name": "Debug MPU Program",
    "program": "${workspaceFolder}/program.s",
    "stopOnEntry": true
}
```

### 2. Breakpoint Support
- Click in editor gutter to set/remove breakpoints
- Conditional breakpoints (future enhancement)
- Hit count breakpoints

### 3. Variable Watch
- Watch window for monitoring variables
- Hover over variables in source to see values
- Locals window showing frame-relative variables

### 4. Call Stack
- Visual call stack with source locations
- Click to navigate to different stack frames

### 5. Memory View
- Hex editor view for memory inspection
- Highlight recently changed memory
- Go to address functionality

### 6. Registers Panel
- Custom panel showing PC, SP, FP
- Visual indicators for flags (N, Z, C, B)

### 7. Disassembly View
- Switch between source and disassembly
- Mixed source/disassembly mode

## Implementation Steps

### Phase 1: Basic DAP Server
1. Implement minimal DAP server in Go
2. Support launch, breakpoints, stepping
3. Basic variable inspection

### Phase 2: VSCode Extension
1. Create extension scaffold
2. Implement debug adapter client
3. Add basic UI integration

### Phase 3: Enhanced Features
1. Memory view webview panel
2. Register view panel
3. Disassembly view

### Phase 4: Advanced Features
1. Conditional breakpoints
2. Watch expressions
3. REPL console

## Example DAP Messages

### SetBreakpoints Request
```json
{
    "command": "setBreakpoints",
    "arguments": {
        "source": {
            "path": "/path/to/program.s"
        },
        "breakpoints": [
            { "line": 10 },
            { "line": 25 }
        ]
    }
}
```

### StackTrace Response
```json
{
    "stackFrames": [
        {
            "id": 1,
            "name": "main",
            "source": {
                "path": "/path/to/program.s"
            },
            "line": 15,
            "column": 0
        }
    ]
}
```

### Variables Response
```json
{
    "variables": [
        {
            "name": "r1",
            "value": "0x001e",
            "type": "word",
            "memoryReference": "0x0010"
        }
    ]
}
```

## Web-based Alternative

For a web-based debugger, consider:

1. **WebSocket API**: Replace DAP with WebSocket for browser communication
2. **Web Components**: Build interactive UI components
3. **Monaco Editor**: Use VSCode's editor component for source viewing
4. **State Synchronization**: Keep debugger state synchronized with UI

## Benefits

- **Familiar Interface**: Developers can use VSCode's debugging UI
- **Standard Protocol**: DAP is well-documented and supported
- **Extensible**: Easy to add new features
- **Cross-platform**: Works on all platforms VSCode supports

## Next Steps

1. Create proof-of-concept DAP server
2. Build minimal VSCode extension
3. Test with real MPU programs
4. Iterate based on user feedback