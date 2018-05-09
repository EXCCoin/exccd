// Package rpctest provides a exccd-specific RPC testing harness crafting and
// executing integration tests by driving a `exccd` instance via the `RPC`
// interface. Each instance of an active harness comes equipped with a simple
// in-memory HD wallet capable of properly syncing to the generated chain,
// creating new addresses, and crafting fully signed transactions paying to an
// arbitrary set of outputs.
//
// This package was designed specifically to act as an RPC testing harness for
// `exccd`. However, the constructs presented are general enough to be adapted to
// any project wishing to programmatically drive a `exccd` instance of its
// systems/integration tests.
package rpctest
