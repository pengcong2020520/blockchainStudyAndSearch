package server

func (srv *Server) run() {
	srv.log.Info("Started P2P networking", "self", srv.localnode.Node().URLv4())
	defer srv.loopWG.Done()
	defer srv.nodedb.Close()
	defer srv.discmix.Close()
	defer srv.dialsched.stop()

	var (
		peers        = make(map[enode.ID]*Peer)
		inboundCount = 0
		trusted      = make(map[enode.ID]bool, len(srv.TrustedNodes))
	)
	// Put trusted nodes into a map to speed up checks.
	// Trusted peers are loaded on startup or added via AddTrustedPeer RPC.
	for _, n := range srv.TrustedNodes {
		trusted[n.ID()] = true
	}

running:
	for {
		select {
		case <-srv.quit:
			// The server was stopped. Run the cleanup logic.
			break running

		case n := <-srv.addtrusted:
			// This channel is used by AddTrustedPeer to add a node
			// to the trusted node set.
			srv.log.Trace("Adding trusted node", "node", n)
			trusted[n.ID()] = true
			if p, ok := peers[n.ID()]; ok {
				p.rw.set(trustedConn, true)
			}

		case n := <-srv.removetrusted:
			// This channel is used by RemoveTrustedPeer to remove a node
			// from the trusted node set.
			srv.log.Trace("Removing trusted node", "node", n)
			delete(trusted, n.ID())
			if p, ok := peers[n.ID()]; ok {
				p.rw.set(trustedConn, false)
			}

		case op := <-srv.peerOp:
			// This channel is used by Peers and PeerCount.
			op(peers)
			srv.peerOpDone <- struct{}{}

		case c := <-srv.checkpointPostHandshake:
			// A connection has passed the encryption handshake so
			// the remote identity is known (but hasn't been verified yet).
			if trusted[c.node.ID()] {
				// Ensure that the trusted flag is set before checking against MaxPeers.
				c.flags |= trustedConn
			}
			// TODO: track in-progress inbound node IDs (pre-Peer) to avoid dialing them.
			c.cont <- srv.postHandshakeChecks(peers, inboundCount, c)

		case c := <-srv.checkpointAddPeer:
			// At this point the connection is past the protocol handshake.
			// Its capabilities are known and the remote identity is verified.
			err := srv.addPeerChecks(peers, inboundCount, c)
			if err == nil {
				// The handshakes are done and it passed all checks.
				p := srv.launchPeer(c)
				peers[c.node.ID()] = p
				srv.log.Debug("Adding p2p peer", "peercount", len(peers), "id", p.ID(), "conn", c.flags, "addr", p.RemoteAddr(), "name", truncateName(c.name))
				srv.dialsched.peerAdded(c)
				if p.Inbound() {
					inboundCount++
				}
			}
			c.cont <- err

		case pd := <-srv.delpeer:
			// A peer disconnected.
			d := common.PrettyDuration(mclock.Now() - pd.created)
			delete(peers, pd.ID())
			srv.log.Debug("Removing p2p peer", "peercount", len(peers), "id", pd.ID(), "duration", d, "req", pd.requested, "err", pd.err)
			srv.dialsched.peerRemoved(pd.rw)
			if pd.Inbound() {
				inboundCount--
			}
		}
	}

	srv.log.Trace("P2P networking is spinning down")

	// Terminate discovery. If there is a running lookup it will terminate soon.
	if srv.ntab != nil {
		srv.ntab.Close()
	}
	if srv.DiscV5 != nil {
		srv.DiscV5.Close()
	}
	// Disconnect all peers.
	for _, p := range peers {
		p.Disconnect(DiscQuitting)
	}
	// Wait for peers to shut down. Pending connections and tasks are
	// not handled here and will terminate soon-ish because srv.quit
	// is closed.
	for len(peers) > 0 {
		p := <-srv.delpeer
		p.log.Trace("<-delpeer (spindown)")
		delete(peers, p.ID())
	}
}
