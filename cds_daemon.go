
	listen, err := net.Listen(TYPE, HOST+":"+PORT)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	// close listener
	defer listen.Close()
	for {
		cds_util.Println("Aguardando conexao...")
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}


		if err := cds_util.ConfigureKeepAlive(conn); err != nil {
			cds_util.Println("Falha ao definir KeepAlive...")
			conn.Close()
			continue
		}

		rmt := conn.RemoteAddr().String()

		log.Printf("Efetuado conexao com cliente(Remote Host=%s).\n", rmt)
		cds_util.Println("Inciando sessao multiplex.")
		go handleRequest(conn)

		cds_util.Println("Encerrada sessao do multiplex.")
	}
}

