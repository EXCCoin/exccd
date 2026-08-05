// Copyright (c) 2016 The btcsuite developers
// Copyright (c) 2017-2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package connmgr

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (

	// maxRetryDuration is the max duration of time retrying of a persistent
	// connection is allowed to grow to.  This is necessary since the retry
	// logic uses a backoff mechanism which increases the interval base times
	// the number of retries that have been done.
	maxRetryDuration = time.Minute * 5
)

const (
	// MaxPersistent is the maximum number of persistent connections that can be
	// registered.  Persistent connections do not count towards the normal
	// connection limit.
	MaxPersistent = 8

	// maxFailedAttempts is the maximum number of successive failed connection
	// attempts after which network failure is assumed and new connections will
	// be delayed by the configured retry duration.
	maxFailedAttempts = 25

	// defaultRetryDuration is the default duration of time for retrying
	// persistent connections.
	defaultRetryDuration = time.Second * 5

	// defaultTargetOutbound is the default number of outbound connections to
	// maintain.
	defaultTargetOutbound = uint32(8)

	// defaultMaxNormalConns is the default maximum number of normal inbound,
	// outbound, and pending connections.
	defaultMaxNormalConns = uint32(125)
)

// ConnState represents the state of the requested connection.
type ConnState uint32

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.
const (
	ConnPending ConnState = iota
	ConnEstablished
	ConnDisconnected
	ConnFailed
	ConnCanceled
)

// ConnReq is the connection request to a network address. If permanent, the
// connection will be retried on disconnection.
type ConnReq struct {
	// The following variables must only be used atomically.
	//
	// id is the unique identifier for this connection request.
	//
	// state is the current connection state for this connection request.
	id    uint64
	state uint32

	// The following fields are owned by the connection handler and must not
	// be accessed outside of it.
	//
	// retryCount is the number of times a permanent connection request that
	// fails to connect has been retried since the last successful connection.
	//
	// conn is the underlying network connection.  It will be nil before a
	// connection has been established.
	retryCount uint32
	conn       net.Conn

	// reservation owns admission resources for the current connection attempt.
	reservationMtx sync.Mutex
	reservation    *connReservation

	// Addr is the address to connect to.
	Addr net.Addr

	// Permanent specifies whether or not the connection request represents what
	// should be treated as a permanent connection, meaning the connection
	// manager will try to always maintain the connection including retries with
	// increasing backoff timeouts.
	Permanent bool
}

// setReservation assigns admission resources to the current attempt.  It
// returns false and releases them when the request was canceled concurrently.
func (c *ConnReq) setReservation(r *connReservation) bool {
	c.reservationMtx.Lock()
	if c.State() == ConnCanceled {
		c.reservationMtx.Unlock()
		r.release()
		return false
	}
	old := c.reservation
	c.reservation = r
	c.reservationMtx.Unlock()
	if old != nil {
		old.release()
	}
	return true
}

// releaseReservation releases admission resources for the current attempt.
func (c *ConnReq) releaseReservation() {
	c.reservationMtx.Lock()
	r := c.reservation
	c.reservation = nil
	c.reservationMtx.Unlock()
	if r != nil {
		r.release()
	}
}

// updateState updates the state of the connection request.
func (c *ConnReq) updateState(state ConnState) {
	atomic.StoreUint32(&c.state, uint32(state))
}

// ID returns a unique identifier for the connection request.
func (c *ConnReq) ID() uint64 {
	return atomic.LoadUint64(&c.id)
}

// State is the connection state of the requested connection.
func (c *ConnReq) State() ConnState {
	return ConnState(atomic.LoadUint32(&c.state))
}

// String returns a human-readable string for the connection request.
func (c *ConnReq) String() string {
	if c.Addr == nil || c.Addr.String() == "" {
		return fmt.Sprintf("reqid %d", atomic.LoadUint64(&c.id))
	}
	return fmt.Sprintf("%s (reqid %d)", c.Addr, atomic.LoadUint64(&c.id))
}

// Config holds the configuration options related to the connection manager.
type Config struct {
	// Listeners defines a slice of listeners for which the connection
	// manager will take ownership of and accept connections.  When a
	// connection is accepted, the OnAccept handler will be invoked with the
	// connection.  Since the connection manager takes ownership of these
	// listeners, they will be closed when the connection manager is
	// stopped.
	//
	// This field will not have any effect if the OnAccept field is not
	// also specified.  It may be nil if the caller does not wish to listen
	// for incoming connections.
	Listeners []net.Listener

	// OnAccept is a callback that is fired when an inbound connection is
	// accepted.  It is the caller's responsibility to close the connection.
	// Failure to close the connection will result in the connection manager
	// believing the connection is still active and thus have undesirable
	// side effects such as still counting toward maximum connection limits.
	//
	// This field will not have any effect if the Listeners field is not
	// also specified since there couldn't possibly be any accepted
	// connections in that case.
	OnAccept func(net.Conn)

	// MaxNormalConns is the maximum number of normal inbound, outbound, and
	// pending connections.  Persistent and whitelisted connections are exempt.
	// Defaults to 125.
	MaxNormalConns uint32

	// MaxConnsPerHost is the maximum number of pending and active connections
	// with the same host.  Zero disables the limit.  Whitelisted and loopback
	// connections are exempt.
	MaxConnsPerHost uint32

	// IsWhitelisted reports whether an address is exempt from normal and
	// per-host connection limits.
	IsWhitelisted func(net.Addr) bool

	// TargetOutbound is the number of outbound network connections to
	// maintain. Defaults to 8.
	TargetOutbound uint32

	// RetryDuration is the duration to wait before retrying connection
	// requests. Defaults to 5s.
	RetryDuration time.Duration

	// OnConnection is a callback that is fired when a new outbound
	// connection is established.
	OnConnection func(*ConnReq, net.Conn)

	// OnDisconnection is a callback that is fired when an outbound
	// connection is disconnected.
	OnDisconnection func(*ConnReq)

	// GetNewAddress is a way to get an address to make a network connection
	// to.  If nil, no new connections will be made automatically.
	GetNewAddress func() (net.Addr, error)

	// OutboundGroup returns the network group for an address.  Automatic
	// outbound selection strongly prefers at most one pending or active
	// connection per group.  Group limiting is disabled when nil.
	OutboundGroup func(net.Addr) string

	// Dial connects to the address on the named network. Either Dial or
	// DialAddr need to be specified (but not both).
	Dial func(ctx context.Context, network, addr string) (net.Conn, error)

	// DialAddr is an alternative to Dial which receives a full net.Addr instead
	// of just the protocol family and address. Either DialAddr or Dial need
	// to be specified (but not both).
	DialAddr func(context.Context, net.Addr) (net.Conn, error)

	// Timeout specifies the amount of time to wait for a connection
	// to complete before giving up.
	Timeout time.Duration
}

// registerPending is used to register a pending connection attempt. By
// registering pending connection attempts we allow callers to cancel pending
// connection attempts before they're successful or in the case they're no
// longer wanted.
type registerPending struct {
	c    *ConnReq
	done chan error
}

// handleConnected is used to queue a successful connection.
type handleConnected struct {
	c    *ConnReq
	conn net.Conn
}

// handleDisconnected is used to remove a connection.
type handleDisconnected struct {
	id    uint64
	retry bool
}

// handleFailed is used to remove a pending connection.
type handleFailed struct {
	c   *ConnReq
	err error
}

// handleCancelPending is used to remove failing connections from retries.
type handleCancelPending struct {
	addr net.Addr
	done chan error
}

// handleForEachConnReq is used to iterate all known connection requests to
// include pending ones.
type handleForEachConnReq struct {
	f    func(c *ConnReq) error
	done chan error
}

// ConnManager provides a manager to handle network connections.
type ConnManager struct {
	// The following variables must only be used atomically.
	//
	// connReqCount is the number of connection requests that have been made and
	// is primarily used to assign unique connection request IDs.
	connReqCount uint64

	// The following fields are used for lifecycle management of the connection
	// manager.
	wg   sync.WaitGroup
	quit chan struct{}

	// cfg specifies the configuration of the connection manager and is set at
	// creating time and treated as immutable after that.
	cfg Config

	// failedAttempts tracks the total number of failed oubound connection
	// attempts since the last successful connection made by the connection
	// manager.  It is primarily used to detect network outages in order to
	// impose a retry timeout on achieving the target number of outbound
	// connections which prevents runaway failed connection attempt churn.
	//
	// This field is owned by the connection handler and must not be accessed
	// outside of it.
	failedAttempts uint64

	// requests is used internally to interact with the connection handler
	// goroutine.
	requests chan interface{}

	// normalSem limits pending and active normal connections.
	normalSem semaphore

	// hostCounts tracks pending and active connections by host.
	hostMtx    sync.Mutex
	hostCounts map[string]uint32

	// outboundGroups tracks pending and active outbound network groups.
	outboundGroups *outboundGroupState

	// inboundLimiter rate limits inbound groups and applies additional pressure
	// when the aggregate admitted connection rate indicates a flood.
	inboundLimiter *inboundRateLimiter
}

// handleFailedConn handles a connection failed due to a disconnect or any
// other failure. If permanent, it retries the connection after the configured
// retry duration. Otherwise, if required, it makes a new connection request.
// After maxFailedConnectionAttempts new connections will be retried after the
// configured retry duration.
func (cm *ConnManager) handleFailedConn(ctx context.Context, c *ConnReq) {
	// Ignore during shutdown.
	if ctx.Err() != nil {
		return
	}

	if c.Permanent {
		c.retryCount++
		d := time.Duration(c.retryCount) * cm.cfg.RetryDuration
		if d > maxRetryDuration {
			d = maxRetryDuration
		}
		log.Debugf("Retrying connection to %v in %v", c, d)
		go func() {
			select {
			case <-time.After(d):
				cm.Connect(ctx, c)
			case <-cm.quit:
			}
		}()
	} else if cm.cfg.GetNewAddress != nil {
		cm.failedAttempts++
		if cm.failedAttempts >= maxFailedAttempts {
			log.Debugf("Max failed connection attempts reached: [%d] "+
				"-- retrying connection in: %v", maxFailedAttempts,
				cm.cfg.RetryDuration)
			go func() {
				select {
				case <-time.After(cm.cfg.RetryDuration):
					cm.newConnReq(ctx)
				case <-cm.quit:
				}
			}()
		} else {
			go cm.newConnReq(ctx)
		}
	}
}

// connHandler handles all connection related requests.  It must be run as a
// goroutine.
//
// The connection handler makes sure that we maintain a pool of active outbound
// connections so that we remain connected to the network.  Connection requests
// are processed and mapped by their assigned ids.
func (cm *ConnManager) connHandler(ctx context.Context) {
	var (
		// pending holds all registered conn requests that have yet to
		// succeed.
		pending = make(map[uint64]*ConnReq)

		// conns represents the set of all actively connected peers.
		conns = make(map[uint64]*ConnReq, cm.cfg.TargetOutbound)

		// numPersistent includes pending, established, and retrying persistent
		// requests.  It is intentionally independent of the normal limit.
		numPersistent int
	)

out:
	for {
		select {
		case req := <-cm.requests:
			switch msg := req.(type) {
			case registerPending:
				connReq := msg.c
				if connReq.Permanent {
					if numPersistent >= MaxPersistent {
						connReq.updateState(ConnFailed)
						msg.done <- MakeError(ErrMaxPersistent, fmt.Sprintf(
							"a maximum of %d persistent connections is allowed",
							MaxPersistent))
						continue
					}

					host, limited, err := cm.limitedHostKey(connReq.Addr)
					if err != nil {
						connReq.updateState(ConnFailed)
						msg.done <- err
						continue
					}
					if limited {
						count := 0
						countHost := func(req *ConnReq) {
							if !req.Permanent {
								return
							}
							reqHost, _, err := addrHostKey(req.Addr)
							if err == nil && reqHost == host {
								count++
							}
						}
						for _, req := range pending {
							countHost(req)
						}
						for _, req := range conns {
							countHost(req)
						}
						if uint32(count) >= cm.cfg.MaxConnsPerHost {
							connReq.updateState(ConnFailed)
							msg.done <- MakeError(ErrMaxPersistentPerHost,
								fmt.Sprintf("a maximum of %d persistent connections per host is allowed",
									cm.cfg.MaxConnsPerHost))
							continue
						}
					}
					numPersistent++
				}
				connReq.updateState(ConnPending)
				pending[msg.c.id] = connReq
				msg.done <- nil

			case handleConnected:
				connReq := msg.c
				if _, ok := pending[connReq.id]; !ok {
					if msg.conn != nil {
						msg.conn.Close()
					}
					connReq.releaseReservation()
					log.Debugf("Ignoring connection for "+
						"canceled connreq=%v", connReq)
					continue
				}

				connReq.updateState(ConnEstablished)
				connReq.conn = msg.conn
				conns[connReq.id] = connReq
				log.Debugf("Connected to %v", connReq)
				connReq.retryCount = 0
				cm.failedAttempts = 0

				delete(pending, connReq.id)

				if cm.cfg.OnConnection != nil {
					go cm.cfg.OnConnection(connReq, msg.conn)
				}

			case handleDisconnected:
				connReq, ok := conns[msg.id]
				if !ok {
					connReq, ok = pending[msg.id]
					if !ok {
						log.Errorf("Unknown connid=%d",
							msg.id)
						continue
					}

					// Pending connection was found, remove
					// it from pending map if we should
					// ignore a later, successful
					// connection.
					connReq.updateState(ConnCanceled)
					log.Debugf("Canceling: %v", connReq)
					delete(pending, msg.id)
					if connReq.Permanent {
						numPersistent--
					}
					connReq.releaseReservation()
					continue
				}

				// An existing connection was located, mark as
				// disconnected and execute disconnection
				// callback.
				log.Debugf("Disconnected from %v", connReq)
				delete(conns, msg.id)

				if connReq.conn != nil {
					connReq.conn.Close()
				}
				connReq.releaseReservation()

				if cm.cfg.OnDisconnection != nil {
					go cm.cfg.OnDisconnection(connReq)
				}

				// All internal state has been cleaned up, if
				// this connection is being removed, we will
				// make no further attempts with this request.
				if !msg.retry {
					if connReq.Permanent {
						numPersistent--
					}
					connReq.updateState(ConnDisconnected)
					continue
				}

				// Otherwise, we will attempt a reconnection if
				// we do not have enough peers, or if this is a
				// persistent peer. The connection request is
				// re added to the pending map, so that
				// subsequent processing of connections and
				// failures do not ignore the request.
				if connReq.Permanent {
					connReq.updateState(ConnPending)
					log.Debugf("Reconnecting to %v",
						connReq)
					pending[msg.id] = connReq
					cm.handleFailedConn(ctx, connReq)
				} else if uint32(len(conns)) < cm.cfg.TargetOutbound {
					connReq.updateState(ConnDisconnected)
					cm.handleFailedConn(ctx, connReq)
				}

			case handleFailed:
				connReq := msg.c
				if _, ok := pending[connReq.id]; !ok {
					log.Debugf("Ignoring connection for "+
						"canceled conn req: %v", connReq)
					continue
				}

				connReq.updateState(ConnFailed)
				connReq.releaseReservation()
				log.Debugf("Failed to connect to %v: %v",
					connReq, msg.err)
				if !connReq.Permanent && cm.cfg.GetNewAddress != nil {
					delete(pending, connReq.id)
				}
				cm.handleFailedConn(ctx, connReq)

			case handleCancelPending:
				pendingAddr := msg.addr.String()
				var idToRemove uint64
				var connReq *ConnReq
				for id, req := range pending {
					if req == nil || req.Addr == nil {
						continue
					}
					if pendingAddr == req.Addr.String() {
						idToRemove, connReq = id, req
						break
					}
				}
				if connReq != nil {
					delete(pending, idToRemove)
					if connReq.Permanent {
						numPersistent--
					}
					connReq.updateState(ConnCanceled)
					connReq.releaseReservation()
					log.Debugf("Canceled pending connection to %v", msg.addr)
					msg.done <- nil
				} else {
					msg.done <- fmt.Errorf("no pending connection to %v", msg.addr)
				}

			case handleForEachConnReq:
				var err error
				for _, connReq := range pending {
					err = msg.f(connReq)
					if err != nil {
						break
					}
				}
				if err != nil {
					msg.done <- err
					continue
				}
				for _, connReq := range conns {
					err = msg.f(connReq)
					if err != nil {
						break
					}
				}
				msg.done <- err
			}

		case <-ctx.Done():
			break out
		}
	}

	// Release all admission resources and close active connections on shutdown.
	for _, connReq := range pending {
		connReq.updateState(ConnCanceled)
		connReq.releaseReservation()
	}
	for _, connReq := range conns {
		if connReq.conn != nil {
			connReq.conn.Close()
		}
		connReq.releaseReservation()
	}

	cm.wg.Done()
	log.Trace("Connection handler done")
}

// registerPendingReq registers a new connection request and returns any
// admission error reported by the connection handler.
func (cm *ConnManager) registerPendingReq(c *ConnReq) error {
	done := make(chan error, 1)
	select {
	case cm.requests <- registerPending{c, done}:
	case <-cm.quit:
		return context.Canceled
	}

	select {
	case err := <-done:
		return err
	case <-cm.quit:
		return context.Canceled
	}
}

// newConnReq creates a new connection request and connects to the
// corresponding address.
func (cm *ConnManager) newConnReq(ctx context.Context) {
	// Ignore during shutdown.
	if ctx.Err() != nil {
		return
	}

	c := &ConnReq{}
	atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))

	// Wait for room before registering another automatic connection so pending
	// and active attempts are both included in the overall limit.
	r, err := cm.reserveOutbound(ctx, c, true)
	if err != nil {
		return
	}
	if !c.setReservation(r) {
		return
	}

	// Submit a request of a pending connection attempt to the connection
	// manager. By registering the id before the connection is even
	// established, we'll be able to later cancel the connection via the
	// Remove method.
	if err := cm.registerPendingReq(c); err != nil {
		c.releaseReservation()
		return
	}

	addr, releaseGroup, err := cm.pickOutboundAddr()
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
		}
		return
	}

	c.Addr = addr
	r.add(releaseGroup)
	releaseHost, err := cm.reserveHost(addr)
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
			c.releaseReservation()
		}
		return
	}
	r.add(releaseHost)

	cm.dial(ctx, c)
}

// Connect assigns an id and dials a connection to the address of the connection
// request using the provided context and the dial function configured when
// initially creating the connection manager.
//
// The connection attempt will be ignored if the connection manager has been
// shutdown by canceling the lifecycle context the Run method was invoked with
// or the provided connection request is already in the failed state.
//
// Note that the context parameter to this function and the lifecycle context
// may be independent.
func (cm *ConnManager) Connect(ctx context.Context, c *ConnReq) {
	// Ignore during shutdown and when caller provided context is already
	// canceled.
	select {
	case <-cm.quit:
		return
	default:
	}
	if ctx.Err() != nil {
		return
	}

	// During the time we wait for retry there is a chance that this
	// connection was already cancelled.
	if c.State() == ConnCanceled {
		log.Debugf("Ignoring connect for canceled connreq=%v", c)
		return
	}

	if atomic.LoadUint64(&c.id) == 0 {
		atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))

		// Submit a request of a pending connection attempt to the
		// connection manager. By registering the id before the
		// connection is even established, we'll be able to later
		// cancel the connection via the Remove method.
		if err := cm.registerPendingReq(c); err != nil {
			return
		}
	}

	r, err := cm.reserveOutbound(ctx, c, false)
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
		}
		return
	}
	if !c.setReservation(r) {
		return
	}
	cm.dial(ctx, c)
}

// ConnectAsync registers a new connection request, performs admission checks,
// and starts the dial asynchronously.  Admission errors are returned before
// the method returns so callers do not report false success.
func (cm *ConnManager) ConnectAsync(ctx context.Context, c *ConnReq) error {
	select {
	case <-cm.quit:
		return context.Canceled
	default:
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.State() == ConnCanceled {
		return context.Canceled
	}
	if c.ID() != 0 {
		return errors.New("connection request is already registered")
	}

	// Persistent requests must be registered first because their limit covers
	// the request throughout retries, not only an individual dial attempt.
	if c.Permanent {
		atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))
		if err := cm.registerPendingReq(c); err != nil {
			return err
		}

		r, err := cm.reserveOutbound(ctx, c, false)
		if err != nil {
			// The persistent request was accepted.  Treat a temporary lack of a
			// host slot like a failed attempt so it remains registered and retries.
			select {
			case cm.requests <- handleFailed{c, err}:
			case <-cm.quit:
				return context.Canceled
			}
			return nil
		}
		if !c.setReservation(r) {
			return context.Canceled
		}
		go cm.dial(ctx, c)
		return nil
	}

	// Reserve normal and per-host capacity before registering the request so a
	// rejected RPC request never leaves a stale pending entry.
	r, err := cm.reserveOutbound(ctx, c, false)
	if err != nil {
		c.updateState(ConnFailed)
		return err
	}
	if !c.setReservation(r) {
		return context.Canceled
	}
	atomic.StoreUint64(&c.id, atomic.AddUint64(&cm.connReqCount, 1))
	if err := cm.registerPendingReq(c); err != nil {
		c.releaseReservation()
		return err
	}

	go cm.dial(ctx, c)
	return nil
}

// dial performs an outbound connection attempt after admission resources have
// been reserved and the request has been registered as pending.
func (cm *ConnManager) dial(ctx context.Context, c *ConnReq) {

	log.Debugf("Attempting to connect to %v", c)

	if cm.cfg.Timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, cm.cfg.Timeout)
		defer cancel()
	}
	var conn net.Conn
	var err error
	if cm.cfg.Dial != nil {
		conn, err = cm.cfg.Dial(ctx, c.Addr.Network(), c.Addr.String())
	} else {
		conn, err = cm.cfg.DialAddr(ctx, c.Addr)
	}
	if err != nil {
		select {
		case cm.requests <- handleFailed{c, err}:
		case <-cm.quit:
			c.releaseReservation()
		}
		return
	}

	select {
	case cm.requests <- handleConnected{c, conn}:
	case <-cm.quit:
		conn.Close()
		c.releaseReservation()
	}
}

// Disconnect disconnects the connection corresponding to the given connection
// id. If permanent, the connection will be retried with an increasing backoff
// duration.
func (cm *ConnManager) Disconnect(id uint64) {
	select {
	case cm.requests <- handleDisconnected{id, true}:
	case <-cm.quit:
	}
}

// Remove removes the connection corresponding to the given connection id from
// known connections.
//
// NOTE: This method can also be used to cancel a lingering connection attempt
// that hasn't yet succeeded.
func (cm *ConnManager) Remove(id uint64) {
	select {
	case cm.requests <- handleDisconnected{id, false}:
	case <-cm.quit:
	}
}

// CancelPending removes the connection corresponding to the given address
// from the list of pending failed connections.
//
// Returns an error if the connection manager is stopped or there is no pending
// connection for the given address.
func (cm *ConnManager) CancelPending(addr net.Addr) error {
	done := make(chan error, 1)
	select {
	case cm.requests <- handleCancelPending{addr, done}:
	case <-cm.quit:
	}

	// Wait for the connection to be removed from the conn manager's
	// internal state.
	select {
	case err := <-done:
		return err
	case <-cm.quit:
		return fmt.Errorf("connection manager stopped")
	}
}

// ForEachConnReq calls the provided function with each connection request known
// to the connection manager, including pending requests.  Returning an error
// from the provided function will stop the iteration early and return said
// error from this function.
//
// This function is safe for concurrent access.
//
// NOTE: This must not call any other connection manager methods during
// iteration or it will result in a deadlock.
func (cm *ConnManager) ForEachConnReq(f func(c *ConnReq) error) error {
	done := make(chan error, 1)
	select {
	case cm.requests <- handleForEachConnReq{f, done}:
	case <-cm.quit:
	}

	select {
	case err := <-done:
		return err
	case <-cm.quit:
		return fmt.Errorf("connection manager stopped")
	}
}

// listenHandler accepts incoming connections on a given listener.  It must be
// run as a goroutine.
func (cm *ConnManager) listenHandler(ctx context.Context, listener net.Listener) {
	log.Infof("Server listening on %s", listener.Addr())
	for ctx.Err() == nil {
		conn, err := listener.Accept()
		if err != nil {
			// Only log the error if not forcibly shutting down.
			if ctx.Err() == nil {
				log.Errorf("Can't accept connection: %v", err)
			}
			continue
		}

		remoteAddr := conn.RemoteAddr()
		_, remoteIP, err := addrHostKey(remoteAddr)
		if err != nil {
			cm.inboundLimiter.logDrop(remoteAddr, "failed to parse address")
			conn.Close()
			continue
		}
		isWhitelisted := cm.isWhitelisted(remoteAddr)
		isLoopback := remoteIP != nil && remoteIP.IsLoopback()

		if !isWhitelisted && !isLoopback {
			if !cm.inboundLimiter.allow(remoteAddr) {
				cm.inboundLimiter.logDrop(remoteAddr, "rate limited")
				conn.Close()
				continue
			}
			if cm.inboundLimiter.shouldDropProbabilistic() {
				cm.inboundLimiter.logDrop(remoteAddr,
					"probabilistically blocked during flood")
				conn.Close()
				continue
			}
		}

		reservation, err := cm.reserveInbound(conn)
		if err != nil {
			cm.inboundLimiter.logDrop(remoteAddr, err.Error())
			conn.Close()
			continue
		}
		go cm.cfg.OnAccept(&limitedConn{Conn: conn, reservation: reservation})
	}

	cm.wg.Done()
	log.Tracef("Listener handler done for %s", listener.Addr())
}

// Run starts the connection manager along with its configured listeners and
// begins connecting to the network.  It blocks until the provided context is
// cancelled.
func (cm *ConnManager) Run(ctx context.Context) {
	log.Trace("Starting connection manager")

	// Start the connection handler goroutine.
	cm.wg.Add(1)
	go cm.connHandler(ctx)

	// Start all the listeners so long as the caller requested them and provided
	// a callback to be invoked when connections are accepted.
	var listeners []net.Listener
	if cm.cfg.OnAccept != nil {
		listeners = cm.cfg.Listeners
	}
	for _, listener := range cm.cfg.Listeners {
		cm.wg.Add(1)
		go cm.listenHandler(ctx, listener)
	}

	// Start enough outbound connections to reach the target number when not
	// in manual connect mode.
	if cm.cfg.GetNewAddress != nil {
		curConnReqCount := atomic.LoadUint64(&cm.connReqCount)
		for i := curConnReqCount; i < uint64(cm.cfg.TargetOutbound); i++ {
			go cm.newConnReq(ctx)
		}
	}

	// Shutdown the connection manager when the context is canceled.
	cm.wg.Add(1)
	go func(ctx context.Context, listeners []net.Listener) {
		<-ctx.Done()
		close(cm.quit)

		// Stop all the listeners on shutdown.  There will not be any
		// listeners if listening is disabled.
		for _, listener := range listeners {
			// Ignore the error since this is shutdown and there is no way
			// to recover anyways.
			_ = listener.Close()
		}

		cm.wg.Done()
	}(ctx, listeners)

	cm.wg.Wait()
	log.Trace("Connection manager stopped")
}

// New returns a new connection manager with the provided configuration.
//
// Use Run to start listening and/or connecting to the network.
func New(cfg *Config) (*ConnManager, error) {
	if cfg.Dial == nil && cfg.DialAddr == nil {
		return nil, MakeError(ErrDialNil, "dial cannot be nil")
	}
	if cfg.Dial != nil && cfg.DialAddr != nil {
		return nil, MakeError(ErrBothDialsFilled,
			"cannot specify both Dial and DialAddr")
	}
	// Default to sane values
	if cfg.RetryDuration <= 0 {
		cfg.RetryDuration = defaultRetryDuration
	}
	if cfg.TargetOutbound == 0 {
		cfg.TargetOutbound = defaultTargetOutbound
	}
	if cfg.MaxNormalConns == 0 {
		cfg.MaxNormalConns = defaultMaxNormalConns
	}
	if cfg.TargetOutbound > cfg.MaxNormalConns {
		cfg.TargetOutbound = cfg.MaxNormalConns
	}
	outboundGroups, err := newOutboundGroupState()
	if err != nil {
		return nil, MakeError(ErrCSPRNG, "unable to initialize connection manager randomness")
	}
	cm := ConnManager{
		cfg:            *cfg, // Copy so caller can't mutate
		requests:       make(chan interface{}),
		quit:           make(chan struct{}),
		normalSem:      makeSemaphore(cfg.MaxNormalConns),
		hostCounts:     make(map[string]uint32),
		outboundGroups: outboundGroups,
		inboundLimiter: newInboundRateLimiter(outboundGroups.key),
	}
	return &cm, nil
}
