package main

import (
	"fmt"
	"strings"
	"time"
)

// Mudei as constantes para valores menores para testes.
const RATE_LIMIT = 5
const TEMPO_ESPERA = 30 * time.Second

// Structs: os tipos vêm DEPOIS do nome da variável
type Credenciais struct {
	usuario string
	senha   string
	proxy   string
}

type Worker struct {
	cred      *Credenciais
	rateAtual int
}

func NovoWorker(credenciais string) *Worker {
	// strings.Split precisa do pacote "strings"
	tempCreds := strings.Split(credenciais, ":")

	creds := &Credenciais{
		usuario: tempCreds[0],
		senha:   tempCreds[1],
		proxy:   tempCreds[2],
	}

	return &Worker{
		cred:      creds,
		rateAtual: 0,
	}
}

func verificarRateLimit(worker *Worker, trabalhos chan *Worker, descansos chan *Worker) {
	// Parênteses no if não são necessários
	if worker.rateAtual >= RATE_LIMIT {
		worker.rateAtual = 0
		descansos <- worker
		fmt.Printf("[WORKER %s] Rate limit atingido! Indo descansar...\n", worker.cred.usuario)
	} else {
		trabalhos <- worker // volta ao final da fila de trabalhos
	}
}

// checkarQuemTaMorcegando recebe os workers cansados e põe eles pra dormir
func checkarQuemTaMorcegando(descansos chan *Worker, trabalhos chan *Worker) {
	for worker := range descansos {
		// Criamos uma goroutine para que os descansos sejam independentes
		go func(w *Worker) {
			time.Sleep(TEMPO_ESPERA)
			fmt.Printf("[WORKER %s] Acabou o descanso, voltando ao trabalho...\n", w.cred.usuario)
			trabalhos <- w // Devolve para a fila de trabalho após o tempo
		}(worker)
	}
}

func darTrabalho(trabalhos chan *Worker, descansos chan *Worker) {
	// Esse loop vai bloquear automaticamente e só rodar quando tiver worker no canal
	for worker := range trabalhos {

		// Inicia o trabalho em background para não travar a fila
		go func(w *Worker) {
			fmt.Printf("[WORKER %s] Trabalhando! %s - Rate limit atual: %d\n", w.cred.usuario, time.Now().Format("15:04:05"), w.rateAtual)

			time.Sleep(10 * time.Second) // Simula o tempo de trabalho
			w.rateAtual++

			verificarRateLimit(w, trabalhos, descansos)
		}(worker)

		// Um pequeno delay só para cadenciar a distribuição (opcional)
		time.Sleep(1 * time.Second)
	}
}

func main() {
	credenciaisIniciais := []string{"usuario1:senha1:proxy1", "usuario2:senha2:proxy2", "usuario3:senha3:proxy3"}

	// Como sabemos a quantidade de itens, criamos canais com "buffer" do tamanho exato
	qtd := len(credenciaisIniciais)
	canalTrabalhos := make(chan *Worker, qtd)
	canalDescansos := make(chan *Worker, qtd)

	// Instanciar os workers e colocar todos inicialmente na fila de trabalho
	for _, cred := range credenciaisIniciais {
		worker := NovoWorker(cred)
		canalTrabalhos <- worker
	}

	fmt.Printf("Trabalhadores instanciados: %d\n", qtd)

	// Inicia as esteiras em background
	go checkarQuemTaMorcegando(canalDescansos, canalTrabalhos)
	go darTrabalho(canalTrabalhos, canalDescansos)

	// Trava a main() para o programa não fechar
	select {}
}
