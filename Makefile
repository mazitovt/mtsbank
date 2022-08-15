net = all-network-1

# creates docker network
network:
	bash scripts/docker_network.sh $(net)

start: network
	make -C generator start
	make -C history start
	make -C analysis start

stop:
	make -C generator stop
	make -C history stop
	make -C analysis stop

test:
	make -C generator test


