package hedge

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

func doConfirmLeader(ctx context.Context, op *Op, conn net.Conn, _ string) {
	var sb strings.Builder
	sb.WriteString(op.buildAckReply(nil))
	if hl, _ := op.HasLock(); !hl {
		sb.Reset()
		sb.WriteString("\n")
	}

	b := []byte(sb.String())
	conn.Write(b)
}

func doSend(ctx context.Context, op *Op, conn net.Conn, msg string) {
	var sb strings.Builder
	serr := base64.StdEncoding.EncodeToString([]byte(ErrNoLeader.Error()))
	fmt.Fprintf(&sb, "%s\n", serr)
	if hl, _ := op.HasLock(); hl {
		sb.Reset()
		serr := base64.StdEncoding.EncodeToString([]byte(ErrNoHandler.Error()))
		fmt.Fprintf(&sb, "%s\n", serr)
		if op.fnLeader != nil {
			payload := strings.Split(msg, " ")[1]
			decoded, _ := base64.StdEncoding.DecodeString(payload)
			data := op.fnLdrData
			if data == nil {
				data = op
			}

			r, e := op.fnLeader(data, decoded) // call leader handler
			if e != nil {
				sb.Reset()
				serr := base64.StdEncoding.EncodeToString([]byte(e.Error()))
				fmt.Fprintf(&sb, "%s\n", serr)
			} else {
				br := base64.StdEncoding.EncodeToString([]byte(""))
				if r != nil {
					br = base64.StdEncoding.EncodeToString(r)
				}

				sb.Reset()
				fmt.Fprintf(&sb, "%s %s\n", CmdAck, br)
			}
		}
	}

	b := []byte(sb.String())
	conn.Write(b)
}

func doBroadcast(ctx context.Context, op *Op, conn net.Conn, msg string) {
	var sb strings.Builder
	serr := base64.StdEncoding.EncodeToString([]byte(ErrNoHandler.Error()))
	fmt.Fprintf(&sb, "%s\n", serr)
	if op.fnBroadcast != nil {
		payload := strings.Split(msg, " ")[1]
		decoded, _ := base64.StdEncoding.DecodeString(payload)
		data := op.fnBcData
		if data == nil {
			data = op
		}

		r, e := op.fnBroadcast(data, decoded) // call broadcast handler
		if e != nil {
			sb.Reset()
			serr := base64.StdEncoding.EncodeToString([]byte(e.Error()))
			fmt.Fprintf(&sb, "%s\n", serr)
		} else {
			br := base64.StdEncoding.EncodeToString([]byte(""))
			if r != nil {
				br = base64.StdEncoding.EncodeToString(r)
			}

			sb.Reset()
			fmt.Fprintf(&sb, "%s %s\n", CmdAck, br)
		}
	}

	b := []byte(sb.String())
	conn.Write(b)
}

func doHeartbeat(ctx context.Context, op *Op, conn net.Conn, msg string) {
	var sb strings.Builder
	op.addMember(strings.Split(msg, " ")[1])
	fmt.Fprintf(&sb, "%s\n", op.encodeMembers())
	conn.Write([]byte(sb.String()))
}

func doMembers(ctx context.Context, op *Op, conn net.Conn, msg string) {
	payload := strings.Split(msg, " ")[1]
	decoded, _ := base64.StdEncoding.DecodeString(payload)
	var m map[string]struct{}
	json.Unmarshal(decoded, &m)
	m[op.hostPort] = struct{}{} // just to be sure
	op.setMembers(m)            // then replace my records
	members := op.getMembers()
	mlist := []string{}
	for k := range members {
		mlist = append(mlist, k)
	}

	op.logger.Printf("%v member(s) tracked", len(op.getMembers()))
	reply := op.buildAckReply(nil)
	conn.Write([]byte(reply))
}

func handleMsg(ctx context.Context, op *Op, conn net.Conn) {
	defer conn.Close()
	fns := map[string]func(ctx context.Context, op *Op, conn net.Conn, msg string){
		CmdLeader:          doConfirmLeader, // confirm leader only
		CmdSend + " ":      doSend,          // Send() API
		CmdBroadcast + " ": doBroadcast,     // Broadcast() API
		CmdPing + " ":      doHeartbeat,     // heartbeat
		CmdMembers + " ":   doMembers,       // broadcast online members
	}

	addSpace := func(s string) string {
		var sb strings.Builder
		fmt.Fprintf(&sb, "%s ", s)
		return sb.String()
	}

	for {
		var prefix string
		msg, err := op.recv(conn)
		if err != nil || ctx.Err() != nil {
			return
		}

		switch {
		case msg == CmdPing: // leader asking if we are online (msg has no prefix)
			reply := op.buildAckReply(nil)
			conn.Write([]byte(reply))
			return
		case strings.HasPrefix(msg, CmdLeader):
			prefix = CmdLeader
		case strings.HasPrefix(msg, addSpace(CmdWrite)):
			prefix = addSpace(CmdWrite)
		case strings.HasPrefix(msg, addSpace(CmdSend)):
			prefix = addSpace(CmdSend)
		case strings.HasPrefix(msg, addSpace(CmdBroadcast)):
			prefix = addSpace(CmdBroadcast)
		case strings.HasPrefix(msg, addSpace(CmdPing)):
			prefix = addSpace(CmdPing)
		case strings.HasPrefix(msg, addSpace(CmdMembers)):
			prefix = addSpace(CmdMembers)
		case strings.HasPrefix(msg, addSpace(CmdSemaphore)):
			prefix = addSpace(CmdSemaphore)
		case strings.HasPrefix(msg, addSpace(CmdSemAcquire)):
			prefix = addSpace(CmdSemAcquire)
		case strings.HasPrefix(msg, addSpace(CmdSemRelease)):
			prefix = addSpace(CmdSemRelease)
		default:
			return // do nothing
		}

		fns[prefix](ctx, op, conn, msg)
	}
}
