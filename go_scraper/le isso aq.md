gerado com ia, mas explica bem a mudanca de pools para canais.

# Refatoração de Concorrência: Migração de Mutex/Pool para Canais (Channels)

Esta refatoração altera a arquitetura de gerência de rotinas concorrentes, abandonando `Slices` + `sync.Mutex` em favor de `Channels`, seguindo o princípio da linguagem Go: *"Não se comunique compartilhando memória; em vez disso, compartilhe memória comunicando-se."*

## 1. Por que os Mutexes e a struct de `Pool` foram removidos?

Na versão anterior, a `Pool` utilizava um *slice* (`[]*Worker`) global. Para evitar *Data Races* e vazamento de memória com múltiplas goroutines tentando ler e modificar esse slice ao mesmo tempo, era exigido o uso custoso de travas (`sync.Mutex`). 

Neste novo modelo:
* **Transferência de Posse (Ownership):** O worker deixou de ser uma variável global compartilhada para se tornar um "token" isolado. Quando a função `darTrabalho` lê o worker do canal, **apenas aquela goroutine tem acesso a ele.** Não existindo acesso livre simultâneo por outras funções, o Mutex torna-se desnecessário.
* **Canais são nativamente seguros:** A `Pool` foi substituída por `chan *Worker`. Em Go, diferentemente de slices, canais funcionam como filas nativamente blindadas contra condição de corrida (*thread-safe queues*).
* **Fim do Busy Waiting:** Antes, gastava-se ciclos de CPU com loops infinitos varrendo a lista. Agora, a leitura do canal (`for worker := range canal`) bloqueia automaticamente com custo zero de processamento até que exista de fato um worker aguardando na fila.

## 2. Como o fluxo continua sendo "infinito" sem um `while (true)`?

Sem um laço explícito infinito, o código continua executando indefinidamente pois foi criado um **ciclo contínuo de fluxo de dados** em duas vias:

1. **Injeção Inicial:** Na `main`, todos os workers são instanciados e jogados no canal `canalTrabalhos`.
2. **Consumo Fixo:** O `range` do canal nunca "termina", ele apenas adormece (sleep) quando o canal esvazia e acorda imediatamente quando um novo item chega.
3. **Mecanismo de Reciclagem:** Logo após o trabalhador terminar de atuar na sua *goroutine* e subir sua contagem, ele não é destruído. Se o `Rate Limit` for aprovado, ele sofre uma devolução ao fim da fila original: `trabalhos <- worker`. Caso reprovado, é mandado ao `descansos <- worker`.
4. **Resgate do Cansaço:** A função ouvinte dos exaustos aguarda de forma inteligente até a finalização do relógio para, finalmente, devolvê-lo novamente para `trabalhos <- worker`. 

Dessa forma temos uma esteira circular ininterrupta, altamente performática e limpa de falsos bloqueios (*Deadlocks*).