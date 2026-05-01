RATE_LIMIT := 100
TEMPO_ESPERA := time.Minute(30)

// usuario:senha:proxy
struct Credenciais {
	string usuario
	string senha
	string proxy
	
}

struct Worker {
	Time lastDescanso
	Credenciais *cred
	int rateAtual
}

struct Pool {
	Worker[] *workers
	int numWorkers
	int rateLimit
}

func NovaPool(numWorkers int) {
	return &Pool{
		nil,
		numWorkers
	}
}

func NovoWorker(credenciais string) *Worker {
	tempCreds string[]
	tempCreds = credenciais.split(":")
	creds := &Credenciais{
		tempCreds[0],
		tempCreds[1],
		tempCreds[2]
	}

	return &Worker{
		time.Now(),
		creds,
		0
	}
}

func criarWorkers(string[] credenciais) Worker[]{
	workers Worker[]
	for i, v in credenciais{
		worker := NovoWorker(v)
		workers.append(worker)
	}

	return workers
}

func instanciarWorkersBulk(workers Worker[], pool *Pool){
	for worker : workers {
		pool.workers.append(worker)
		pool.numWorkers++
	}

}

func moverWorker(worker Worker, poolInicio Pool, poolDestino Pool){
	poolInicio.workers.remove(worker)
	poolDestino.workers.append(worker)
}

func verificarRateLimit(worker Worker, poolDescanso Pool, poolTrabalhando Pool){
	if(worker.rateAtual == RATE_LIMIT) {
		worker.rateLimit = 0;
		worker.lastDescanso = time.Now()
		moverWorker(worker, poolTrabalhando, poolDescanso)
	}

}


go func checkarQuemTaMorcegando(poolDescanso Pool, poolTrabalhando Pool) {
	while(true){

		for worker : poolDescanso.workers {
			if(time.Now - worker.lastDescanso >= TEMPO_ESPERA) {
				moverWorker(worker, poolDescanso, poolTrabalhando)

			}

		}

	}
	time.sleep(time.Second(30))
}




credenciaisIniciais := ["usuario1:senha1:proxy1", "usuario2:senha2:proxy2", "usuario3:senha3:proxy3"]

workers := criarWorkers(credenciaisIniciais)

PoolDescanso := NovaPool(NUM_WORKERS)
PoolTrabalhando := NovaPool(NUM_WORKERS)

instanciarWorkersBulk(workers, PoolTrabalhando)

checkarQuemTaMorcegando(PoolDescanso, PoolTrabalhando)
