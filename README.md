# Light Up Davis

A free and open source light control platform by Anna Lynch.
Created to turn Davis Library into a Connect 4 game or maybe Bad Apple?

## What You'll Need
A bunch (6) of *worker* nodes (microcontrollers) each wired up to a handful (7) of lights.
A *controller* node (laptop).
A platform for networking them (k8s over school wifi?, a router?).
Black boxes to obfuscate lights.
5 of the best friends you could ever have in the world.

## Instruction Flow

1. The *controller* transmits each instruction to the responsible controller over HTTP within the cluster.
2. Each *worker* reads the incoming HTTP transmission. The web server component verifies authenticity and safety, then transmits the updates over a buffered channel to the process queue.
3. Each *worker*'s process queue component adds incoming instruction messages to their queue.
4. Each *worker* continuously iterates through its queue, creating a new thread for each message to execute the instruction.
5. Each *worker* executes the instructions per the injected handler function.
