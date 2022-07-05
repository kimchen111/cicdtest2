#!/bin/sh
nft delete table inet qos
nft add table inet qos

nft 'add set inet qos band1_service { type inet_service ; }'
nft 'add set inet qos band2_service { type inet_service ; }'
nft 'add set inet qos band3_service { type inet_service ; }'

nft 'add set inet qos band1_addr { type ipv4_addr ; flags interval; }'
nft 'add set inet qos band2_addr { type ipv4_addr ; flags interval; }'
nft 'add set inet qos band3_addr { type ipv4_addr ; flags interval; }'

nft 'add chain inet qos band1_mark { type filter hook prerouting priority mangle; }'
nft 'add chain inet qos band2_mark { type filter hook prerouting priority mangle; }'
nft 'add chain inet qos band3_mark { type filter hook prerouting priority mangle; }'

nft add rule inet qos band1_mark meta l4proto { tcp, udp } th sport @band1_service meta mark set 0x10 ct mark set mark counter return
nft add rule inet qos band1_mark meta l4proto { tcp, udp } th dport @band1_service meta mark set 0x10 ct mark set mark counter return

nft add rule inet qos band2_mark meta l4proto { tcp, udp } th sport @band2_service meta mark set 0x20 ct mark set mark counter return
nft add rule inet qos band2_mark meta l4proto { tcp, udp } th dport @band2_service meta mark set 0x20 ct mark set mark counter return

nft add rule inet qos band3_mark meta l4proto { tcp, udp } th sport @band3_service meta mark set 0x30 ct mark set mark counter return
nft add rule inet qos band3_mark meta l4proto { tcp, udp } th dport @band3_service meta mark set 0x30 ct mark set mark counter return

nft add rule inet qos band1_mark ip saddr @band1_addr meta mark 0x10 ct mark set mark counter return
nft add rule inet qos band1_mark ip daddr @band1_addr meta mark 0x10 ct mark set mark counter return

nft add rule inet qos band2_mark ip saddr @band2_addr meta mark 0x20 ct mark set mark counter return
nft add rule inet qos band2_mark ip daddr @band2_addr meta mark 0x20 ct mark set mark counter return

nft add rule inet qos band3_mark ip saddr @band3_addr meta mark 0x30 ct mark set mark counter return
nft add rule inet qos band3_mark ip daddr @band3_addr meta mark 0x30 ct mark set mark counter return

nft list table inet qos
