# Nfa Scanner Design

This document explains the nondeterministic finite automaton (Nfa) used by Spot's runtime scanner.
It is intentionally technical, but it starts from first principles.
You should be able to read it without prior automata theory experience and come away understanding:

* What an Nfa is.
* How an Nfa matches text.
* Why Spot uses an Nfa for tokenization.
* How Spot builds an Nfa from DSL token definitions.
* How the Nfa implementation in Spot maps to the code.

# Why?

Spot tokenizes source text from token definitions written in the DSL.
Those token definitions support:

* Character literals.
* String literals.
* Ranges.
* Concatenation.
* Alternation.
* Operator `?`.
* Operator `*`.
* Operator `+`.

That feature set is small, but it is still expressive enough that ad hoc matching code quickly becomes hard to reason
about.

An Nfa gives Spot a precise and uniform way to represent every token definition.
Each DSL expression is compiled into a small graph. The scanner then walks that graph over the input text.

This matters for three reasons:

1. Correctness.
   A single execution model handles all supported token constructs.
2. Simplicity.
   The scanner does not need one matching algorithm for strings, another for repetition, and another for alternation.
3. Extensibility.
   If Spot later adds a Dfa optimization, the Nfa is already the right intermediate representation to convert from.

# What is an Nfa?

An Nfa is a graph of states and transitions.
You can think of it as a machine that answers one question:

> If I start here and read these bytes, can I reach an accepting state?

The important idea is that an Nfa is allowed to be in more than one state at the same time.
That is why it is called nondeterministic. When the graph branches, the machine conceptually follows every branch at
once.

This does not mean Spot launches concurrent goroutines or does anything parallel.
It only means the scanner keeps a set of possible states instead of a single current state.

# States and transitions

In Spot's implementation, a state is a small struct.
At runtime, the scanner only needs a few state kinds:

* A `stateByte`.
* A `stateRange`.
* A `stateEpsilon`.
* A `stateSplit`.
* A `stateAccept`.

These names come directly from the code in `app/scanner/nfa.go`.

## Consuming states

Some states consume one input byte.

### The `stateByte`

This state matches one exact byte.

Example:

```text
'a'
```

Matches only `a`.

### The `stateRange`

This state matches one byte inside an inclusive range.

Example:

```text
'0'..'9'
```

Matches any single digit byte.

## Non-consuming states

Some states do not consume input at all.

### The `stateEpsilon`

An epsilon transition means:

> Move to the next state without reading a byte.

This is useful for connecting fragments together.

### The `stateSplit`

A split state is a branch point with two outgoing epsilon transitions.

It means:

> Continue down either path without consuming input.

This is how alternation and optional/repeating constructs are represented.

### The `stateAccept`

An accept state means:

> A token definition has matched successfully up to this input offset.

In Spot, accept states also store the source-order token index so tie-breaking can follow the DSL rules.

# How the scanner uses an Nfa

The scanner reads the input from left to right.
At each input offset:

1. It starts from the combined Nfa start state.
2. It expands every epsilon transition reachable from that start state.
3. It reads one byte.
4. It advances every state that can consume that byte.
5. It expands epsilon transitions again.
6. It records any accepting states it reaches.
7. It repeats until no states can continue.

At the end of that process, Spot applies the DSL selection rules from `docs/dsl.md`:

1. The longest match wins.
2. If multiple tokens match the same length, the token declared first wins.
3. If no token matches, scanning fails.

This is why the scanner stores a "best match so far" while it is simulating the Nfa.
The machine may have multiple possible matches at the same time, but the scanner only emits the single best one.

# Epsilon closure

The most important NFA concept in Spot's implementation is the epsilon closure.

The epsilon closure of a state is:

> The set of all states reachable from that state by following only epsilon and split transitions.

This matters because a scanner should not stop at a branch point and wait for input if it can move forward without
consuming anything.

For example, consider:

```text
'a'?
```

That expression means:

* Match `a`.
* Or skip it entirely.

Before the scanner reads any input, both possibilities must already be considered.
That is exactly what epsilon closure does.

In Spot:

* Using `buildClosures` computes the closure of every state once when the scanner is created.
* Using `buildClosure` walks through epsilon and split transitions until there is nowhere else to go.

Precomputing closures keeps the scanner loop simpler and avoids recomputing the same graph walk for every input byte.

# Why Spot uses Thompson-style Nfa construction

Spot builds the NFA with Thompson-style construction.
The core idea is simple:

* Every expression compiles to a small fragment.
* Each fragment has one entry state.
* Each fragment has one open end state.
* Larger expressions are formed by connecting smaller fragments.

This style fits the DSL well because every token feature can be expressed through a small number of graph patterns.

The implementation calls these temporary pieces `fragment`.
A fragment has:

* A `start`.
* An `end`.

Those are just indexes into the scanner's state slice.

# How each DSL construct compiles

This section maps DSL constructs to the Nfa builder.

## Character literal

DSL:

```spot
'a'
```

Shape:

```text
[byte 'a'] -> [end]
```

## String literal

DSL:

```spot
"spot"
```

Shape:

```text
[s] -> [p] -> [o] -> [t] -> [end]
```

The string is represented as a concatenation of exact-byte states.

## Range

DSL:

```spot
'0'..'9'
```

Shape:

```text
[range 0-9] -> [end]
```

## Concatenation

DSL:

```spot
letter digit
```

Shape:

```text
[letter fragment] -> [digit fragment]
```

The end of the left fragment is linked to the start of the right fragment.

## Alternation

DSL:

```spot
'a' | 'b'
```

Shape:

```text
          -> [a fragment] -
[split] -                  -> [end]
          -> [b fragment] -
```

The scanner can follow either branch.

## Zero or one

DSL:

```spot
'a'?
```

Shape:

```text
          -> [a fragment] -
[split] -                  -> [end]
          -----------------
```

One branch consumes `a`.
The other skips it.

## Zero or more

DSL:

```spot
'a'*
```

Shape:

```text
          -> [a fragment] -> [split] -
[split] -                              -> [end]
          -----------------------------
```

The first split allows zero repetitions.
The loop split allows either another repetition or exit.

## One or more

DSL:

```spot
'a'+
```

Shape:

```text
[a fragment] -> [split] -
                         -> [end]
               ----------
```

The first repetition is required because the fragment is entered directly.
The split only appears after the first pass through the inner fragment.

# How multiple tokens becomes one machine

Each token definition is compiled independently first.
After that, Spot combines them into one machine by creating a shared start that branches to every token fragment.

If the DSL contains:

```spot
tokens {
    Keyword = "public"
    Identifier = ('a'..'z' | 'A'..'Z')+
}
```

then Spot builds:

```text
                 -> [Keyword fragment] -> [accept Keyword]
[shared start] -
                 -> [Identifier fragment] -> [accept Identifier]
```

This lets the scanner evaluate all token definitions at the same input offset in one pass.

That is the key reason an Nfa fits scanning well:
the DSL says "try every token here", and the combined start state expresses exactly that.

# Concrete Example

Consider this token definition:

```spot
Identifier = ('a'..'z' | 'A'..'Z') ('a'..'z' | 'A'..'Z' | '0'..'9' | '_')*
```

This means:

1. The first byte must be a letter.
2. Every following byte may be a letter, a digit, or `_`.
3. There may be zero or more following bytes.

Spot compiles it conceptually like this:

```text
start
  |
  v
[split first-letter]
  |                |
  v                v
[a-z]            [A-Z]
  \                /
   \              /
    v            v
    [join]
      |
      v
   [split loop] ------------------------.
    |                                   |
    v                                   |
 [split identifier-part]                |
  |      |      |       |               |
  v      v      v       v               |
[a-z]  [A-Z]  [0-9]   [_]               |
  \      |      |      /                |
   \     |      |     /                 |
    \    |      |    /                  |
     v   v      v   v                   |
        [join part] --------------------'
             |
             v
           accept
```

That diagram is simplified, but it captures the control flow:

* The first byte has two possible branches.
* Those branches join.
* Then the machine enters a loop.
* Aach loop iteration allows one of several byte classes.
* The loop may stop at any time and accept.

Now imagine scanning:

```text
alpha2
```

At byte offset `0`:

* The shared start expands to this token's start.
* The machine consumes `a`.
* The scanner records that it can continue.

At each later byte:

* The loop consumes `l`, `p`, `h`, `a`, `2`.
* After each consumed byte, the machine can either continue the loop or stop.

When the scanner reaches the end of `alpha2`, the best recorded accept point is the full six-byte span.

# Why a set of active states is needed?

Many readers expect a scanner to have exactly one current state.
That is true for a Dfa, but not for an Nfa.

In Spot, the active state set is necessary because:

* Alternation creates multiple possible paths.
* Optional expressions create both "take it" and "skip it" paths.
* Repetition creates both "repeat" and "stop" paths.

The implementation therefore keeps slices of active state indexes rather than a single state index.

It also deduplicates those slices.
Without deduplication, the same state could be added repeatedly through different graph paths, which would waste work
and make loops harder to reason about.

# Why the implementation needs deduplication

Two forms of deduplication appear in the code.

## Closure Deduplication

Inside `buildClosure`, the implementation tracks which states were already visited while computing an epsilon closure.
This is required because graphs with repetition naturally create cycles.

Example:

```spot
('a'?)* 'b'
```

The `*` and `?` operators both introduce split states.
Together they create a cycle of epsilon transitions.

Without a `seen` check, closure computation could revisit the same states forever.

## Next-State deduplication

Inside `addClosures`, the implementation tracks which states have already been added to the next active set.
This is required because different active branches may consume the same byte and then reconverge to overlapping
closures.

Example:

```spot
"ab" | "ab"
```

After consuming `a`, both branches are in equivalent positions.
After consuming `b`, both branches reach overlapping closure states.

Deduplication avoids doing the same work twice in later steps.

# How Spot chooses the winning token

Nfas can accept multiple matches at the same time. That is expected.

Spot resolves them with the DSL scanner rules:

1. Prefer the match with the greatest end offset.
2. If the end offsets are equal, prefer the token with the smaller token index.

The smaller token index is just source order from the DSL.
This is why accept states store the token index.
It allows the runtime to preserve the language semantics without extra lookup structures.

# Why Spot starts with an Nfa instead of a Dfa?

A Dfa can be faster at runtime because it has only one current state.
But building and storing a Dfa is more complicated.

Spot currently starts with an Nfa because:

* The DSL feature set maps directly to Thompson construction.
* The implementation is small and explicit.
* Correctness is easier to explain and test.
* A Dfa can still be added later as an optimization stage.

This follows the project principle of choosing the simplest solution that satisfies the current requirements.

The Nfa is not a dead end.
If profiling later shows that runtime scanning is a measurable bottleneck, the combined Nfa can be converted into a Dfa
through subset construction. That optimization would build on the same semantics already defined here.

# How This Maps To The Spot Codebase

The relevant files are:

* `app/compiler/compiler.go`
* `app/ir/program.go`
* `app/scanner/nfa.go`
* `app/scanner/scanner.go`

Their responsibilities are:

* The package `compiler` lowers validated syntax into runtime-oriented token expressions.
* The package `ir` stores the compiled token program.
* The file `scanner/nfa.go` builds the Nfa graph.
* The file `scanner/scanner.go` simulates the machine over input bytes.

The execution flow is:

1. Parse DSL into syntax.
2. Validate syntax.
3. Compile syntax into `ir.Program`.
4. Build an Nfa-backed scanner from `ir.Program`.
5. Scan source text into a token stream.

# Summary

An Nfa is just a graph that represents all allowed matching paths for a token definition.
Spot uses an Nfa because the DSL naturally describes branching and repetition, and the Nfa gives one clear execution
model for all of it.

The important implementation ideas are:

* Each expression compiles into a fragment.
* Fragments are connected with epsilon and split states.
* The scanner keeps a set of active states.
* Epsilon closures are precomputed.
* The scanner records the best accepting match.
* Longest match and source order determine the winning token.

If you understand those points, you understand both the concept and Spot's implementation.
