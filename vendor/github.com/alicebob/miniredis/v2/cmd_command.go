// Command 'COMMAND' from https://redis.io/commands#server

package miniredis

import "github.com/alicebob/miniredis/v2/server"

func commandsCommand(m *Miniredis) {
	_ = m.srv.Register("COMMAND", m.cmdCommand)
}

func (m *Miniredis) cmdCommand(c *server.Peer, cmd string, args []string) {
	// Got from redis 5.0.7 with
	// echo 'COMMAND' | nc redis_addr redis_port
	//
	res := "*200\r\n*6\r\n$12\r\nhincrbyfloat\r\n:4\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$10\r\nxreadgroup\r\n:-7\r\n*3\r\n+write\r\n+noscript\r\n+movablekeys\r\n:1\r\n:1\r\n:1\r\n*6\r\n$10\r\nsdiffstore\r\n:-3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$8\r\nlastsave\r\n:1\r\n*2\r\n+random\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nsetnx\r\n:3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$8\r\nbzpopmax\r\n:-3\r\n*3\r\n+write\r\n+noscript\r\n+fast\r\n:1\r\n:-2\r\n:1\r\n*6\r\n$12\r\npunsubscribe\r\n:-1\r\n*4\r\n+pubsub\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nxack\r\n:-4\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$10\r\npfselftest\r\n:1\r\n*1\r\n+admin\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nsubstr\r\n:4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$8\r\nsmembers\r\n:2\r\n*2\r\n+readonly\r\n+sort_for_script\r\n:1\r\n:1\r\n:1\r\n*6\r\n$11\r\nunsubscribe\r\n:-1\r\n*4\r\n+pubsub\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$11\r\nzinterstore\r\n:-4\r\n*3\r\n+write\r\n+denyoom\r\n+movablekeys\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nstrlen\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\npfmerge\r\n:-2\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$9\r\nrandomkey\r\n:1\r\n*2\r\n+readonly\r\n+random\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nlolwut\r\n:-1\r\n*1\r\n+readonly\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nrpop\r\n:2\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nhkeys\r\n:2\r\n*2\r\n+readonly\r\n+sort_for_script\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nclient\r\n:-2\r\n*2\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nmodule\r\n:-2\r\n*2\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\nslowlog\r\n:-2\r\n*2\r\n+admin\r\n+random\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\ngeohash\r\n:-2\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nlrange\r\n:4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nping\r\n:-1\r\n*2\r\n+stale\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$8\r\nbitcount\r\n:-2\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\npubsub\r\n:-2\r\n*4\r\n+pubsub\r\n+random\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nrole\r\n:1\r\n*3\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nhget\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nobject\r\n:-2\r\n*2\r\n+readonly\r\n+random\r\n:2\r\n:2\r\n:1\r\n*6\r\n$9\r\nzrevrange\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nhincrby\r\n:4\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$9\r\nzlexcount\r\n:4\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nscard\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nappend\r\n:3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nhstrlen\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nconfig\r\n:-2\r\n*4\r\n+admin\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nhset\r\n:-4\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$16\r\nzrevrangebyscore\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nincr\r\n:2\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nsetbit\r\n:4\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$9\r\nrpoplpush\r\n:3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:2\r\n:1\r\n*6\r\n$6\r\nxclaim\r\n:-6\r\n*3\r\n+write\r\n+random\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$11\r\nsinterstore\r\n:-3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$7\r\npublish\r\n:3\r\n*4\r\n+pubsub\r\n+loading\r\n+stale\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nhscan\r\n:-3\r\n*2\r\n+readonly\r\n+random\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nmulti\r\n:1\r\n*2\r\n+noscript\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$3\r\nset\r\n:-3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nlpushx\r\n:-3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$16\r\nzremrangebyscore\r\n:4\r\n*1\r\n+write\r\n:1\r\n:1\r\n:1\r\n*6\r\n$9\r\npexpireat\r\n:3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nhdel\r\n:-3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$12\r\nbgrewriteaof\r\n:1\r\n*2\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\nmigrate\r\n:-6\r\n*3\r\n+write\r\n+random\r\n+movablekeys\r\n:0\r\n:0\r\n:0\r\n*6\r\n$9\r\nreplicaof\r\n:3\r\n*3\r\n+admin\r\n+noscript\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\ntouch\r\n:-2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nxsetid\r\n:3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nbitop\r\n:-4\r\n*2\r\n+write\r\n+denyoom\r\n:2\r\n:-1\r\n:1\r\n*6\r\n$6\r\nswapdb\r\n:3\r\n*2\r\n+write\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nsdiff\r\n:-2\r\n*2\r\n+readonly\r\n+sort_for_script\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$6\r\nlindex\r\n:3\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nwait\r\n:3\r\n*1\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nlrem\r\n:4\r\n*1\r\n+write\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nhsetnx\r\n:4\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$8\r\ngetrange\r\n:4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nhlen\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\npost\r\n:-1\r\n*2\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$9\r\nsismember\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nunwatch\r\n:1\r\n*2\r\n+noscript\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nlpush\r\n:-3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nscan\r\n:-2\r\n*2\r\n+readonly\r\n+random\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nsmove\r\n:4\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:2\r\n:1\r\n*6\r\n$7\r\ncluster\r\n:-2\r\n*1\r\n+admin\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nbgsave\r\n:-1\r\n*2\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\ndump\r\n:2\r\n*2\r\n+readonly\r\n+random\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nlatency\r\n:-2\r\n*4\r\n+admin\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$8\r\nbzpopmin\r\n:-3\r\n*3\r\n+write\r\n+noscript\r\n+fast\r\n:1\r\n:-2\r\n:1\r\n*6\r\n$6\r\ngetbit\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nhgetall\r\n:2\r\n*2\r\n+readonly\r\n+random\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nrename\r\n:3\r\n*1\r\n+write\r\n:1\r\n:2\r\n:1\r\n*6\r\n$9\r\nsubscribe\r\n:-2\r\n*4\r\n+pubsub\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nxdel\r\n:-3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$15\r\nzremrangebyrank\r\n:4\r\n*1\r\n+write\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\ntype\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nscript\r\n:-2\r\n*1\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nhmset\r\n:-4\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nsunion\r\n:-2\r\n*2\r\n+readonly\r\n+sort_for_script\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$4\r\nmget\r\n:-2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$10\r\nbrpoplpush\r\n:4\r\n*3\r\n+write\r\n+denyoom\r\n+noscript\r\n:1\r\n:2\r\n:1\r\n*6\r\n$6\r\ngeoadd\r\n:-5\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\ndecrby\r\n:3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\necho\r\n:2\r\n*1\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\ndbsize\r\n:1\r\n*2\r\n+readonly\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nzcard\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nselect\r\n:2\r\n*2\r\n+loading\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nsadd\r\n:-3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nhost:\r\n:-1\r\n*2\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nsscan\r\n:-3\r\n*2\r\n+readonly\r\n+random\r\n:1\r\n:1\r\n:1\r\n*6\r\n$12\r\ngeoradius_ro\r\n:-6\r\n*2\r\n+readonly\r\n+movablekeys\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nmonitor\r\n:1\r\n*2\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$14\r\nzremrangebylex\r\n:4\r\n*1\r\n+write\r\n:1\r\n:1\r\n:1\r\n*6\r\n$11\r\nsunionstore\r\n:-3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$5\r\nzscan\r\n:-3\r\n*2\r\n+readonly\r\n+random\r\n:1\r\n:1\r\n:1\r\n*6\r\n$9\r\nreadwrite\r\n:1\r\n*1\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nxgroup\r\n:-2\r\n*2\r\n+write\r\n+denyoom\r\n:2\r\n:2\r\n:1\r\n*6\r\n$5\r\nsetex\r\n:4\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nsave\r\n:1\r\n*2\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nhvals\r\n:2\r\n*2\r\n+readonly\r\n+sort_for_script\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nwatch\r\n:-2\r\n*2\r\n+noscript\r\n+fast\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$7\r\nhexists\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\ninfo\r\n:-1\r\n*3\r\n+random\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\npsync\r\n:3\r\n*3\r\n+readonly\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$11\r\nzrangebylex\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nzadd\r\n:-4\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nxlen\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nauth\r\n:2\r\n*4\r\n+noscript\r\n+loading\r\n+stale\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nsrem\r\n:-3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$9\r\ngeoradius\r\n:-6\r\n*2\r\n+write\r\n+movablekeys\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nexec\r\n:1\r\n*2\r\n+noscript\r\n+skip_monitor\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\npfcount\r\n:-2\r\n*1\r\n+readonly\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$7\r\nzpopmin\r\n:-2\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nmove\r\n:3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nxtrim\r\n:-2\r\n*3\r\n+write\r\n+random\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nasking\r\n:1\r\n*1\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\npttl\r\n:2\r\n*3\r\n+readonly\r\n+random\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$11\r\nsrandmember\r\n:-2\r\n*2\r\n+readonly\r\n+random\r\n:1\r\n:1\r\n:1\r\n*6\r\n$8\r\nflushall\r\n:-1\r\n*1\r\n+write\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nsort\r\n:-2\r\n*3\r\n+write\r\n+denyoom\r\n+movablekeys\r\n:1\r\n:1\r\n:1\r\n*6\r\n$3\r\ndel\r\n:-2\r\n*1\r\n+write\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$14\r\nrestore-asking\r\n:-4\r\n*3\r\n+write\r\n+denyoom\r\n+asking\r\n:1\r\n:1\r\n:1\r\n*6\r\n$10\r\npsubscribe\r\n:-2\r\n*4\r\n+pubsub\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\ndecr\r\n:2\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nincrby\r\n:3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$14\r\nzrevrangebylex\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$8\r\nbitfield\r\n:-2\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nexists\r\n:-2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$8\r\nreplconf\r\n:-1\r\n*4\r\n+admin\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\nzincrby\r\n:4\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nblpop\r\n:-3\r\n*2\r\n+write\r\n+noscript\r\n:1\r\n:-2\r\n:1\r\n*6\r\n$4\r\nlpop\r\n:2\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$3\r\nttl\r\n:2\r\n*3\r\n+readonly\r\n+random\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nxread\r\n:-4\r\n*3\r\n+readonly\r\n+noscript\r\n+movablekeys\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nrpush\r\n:-3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$8\r\nzrevrank\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$11\r\nincrbyfloat\r\n:3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nbrpop\r\n:-3\r\n*2\r\n+write\r\n+noscript\r\n:1\r\n:-2\r\n:1\r\n*6\r\n$4\r\nxadd\r\n:-5\r\n*4\r\n+write\r\n+denyoom\r\n+random\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$8\r\nsetrange\r\n:4\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$17\r\ngeoradiusbymember\r\n:-5\r\n*2\r\n+write\r\n+movablekeys\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nunlink\r\n:-2\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$8\r\nexpireat\r\n:3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\ndebug\r\n:-2\r\n*2\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$20\r\ngeoradiusbymember_ro\r\n:-5\r\n*2\r\n+readonly\r\n+movablekeys\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nlset\r\n:4\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nzscore\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nllen\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\ntime\r\n:1\r\n*2\r\n+random\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$8\r\nshutdown\r\n:-1\r\n*4\r\n+admin\r\n+noscript\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\nevalsha\r\n:-3\r\n*2\r\n+noscript\r\n+movablekeys\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nzcount\r\n:4\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nmemory\r\n:-2\r\n*2\r\n+readonly\r\n+random\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nxinfo\r\n:-2\r\n*2\r\n+readonly\r\n+random\r\n:2\r\n:2\r\n:1\r\n*6\r\n$8\r\nxpending\r\n:-3\r\n*2\r\n+readonly\r\n+random\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\neval\r\n:-3\r\n*2\r\n+noscript\r\n+movablekeys\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nxrange\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nrestore\r\n:-4\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nzpopmax\r\n:-2\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nmset\r\n:-3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:-1\r\n:2\r\n*6\r\n$4\r\nspop\r\n:-2\r\n*3\r\n+write\r\n+random\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nltrim\r\n:4\r\n*1\r\n+write\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\nzrank\r\n:3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$9\r\nxrevrange\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$3\r\nget\r\n:2\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nflushdb\r\n:-1\r\n*1\r\n+write\r\n:0\r\n:0\r\n:0\r\n*6\r\n$5\r\nhmget\r\n:-3\r\n*2\r\n+readonly\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nmsetnx\r\n:-3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:-1\r\n:2\r\n*6\r\n$7\r\npersist\r\n:2\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$11\r\nzunionstore\r\n:-4\r\n*3\r\n+write\r\n+denyoom\r\n+movablekeys\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\ncommand\r\n:0\r\n*3\r\n+random\r\n+loading\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$8\r\nrenamenx\r\n:3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:2\r\n:1\r\n*6\r\n$6\r\nzrange\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\npexpire\r\n:3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nkeys\r\n:2\r\n*2\r\n+readonly\r\n+sort_for_script\r\n:0\r\n:0\r\n:0\r\n*6\r\n$4\r\nzrem\r\n:-3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$5\r\npfadd\r\n:-2\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\npsetex\r\n:4\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$13\r\nzrangebyscore\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$4\r\nsync\r\n:1\r\n*3\r\n+readonly\r\n+admin\r\n+noscript\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\npfdebug\r\n:-3\r\n*1\r\n+write\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\ndiscard\r\n:1\r\n*2\r\n+noscript\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$8\r\nreadonly\r\n:1\r\n*1\r\n+fast\r\n:0\r\n:0\r\n:0\r\n*6\r\n$7\r\ngeodist\r\n:-4\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\ngeopos\r\n:-2\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nbitpos\r\n:-3\r\n*1\r\n+readonly\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nsinter\r\n:-2\r\n*2\r\n+readonly\r\n+sort_for_script\r\n:1\r\n:-1\r\n:1\r\n*6\r\n$6\r\ngetset\r\n:3\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nslaveof\r\n:3\r\n*3\r\n+admin\r\n+noscript\r\n+stale\r\n:0\r\n:0\r\n:0\r\n*6\r\n$6\r\nrpushx\r\n:-3\r\n*3\r\n+write\r\n+denyoom\r\n+fast\r\n:1\r\n:1\r\n:1\r\n*6\r\n$7\r\nlinsert\r\n:5\r\n*2\r\n+write\r\n+denyoom\r\n:1\r\n:1\r\n:1\r\n*6\r\n$6\r\nexpire\r\n:3\r\n*2\r\n+write\r\n+fast\r\n:1\r\n:1\r\n:1\r\n"

	c.WriteRaw(res)
}
