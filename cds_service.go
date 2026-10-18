	// Connect to the server
	zrhost := zcds.Rhost

	if ! strings.Contains(zrhost, ":65008") {
		zrhost = zrhost + ":" + PORT
	}

 	conn, err := net.Dial("tcp", zrhost)
	if err != nil {
		fmt.Println("Error:", err)
		return exit_code
	}
	defer conn.Close()


	// habilitar o keep alive na sessao
	if err := cds_util.ConfigureKeepAlive(conn); err != nil {
		fmt.Println("Erro ao configurar keep-alive:", err)
		conn.Close()
		return exit_code
	}



	log.Printf("Efetuada conexao com daemon.\n")


