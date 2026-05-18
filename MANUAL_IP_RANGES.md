# Manual: adicionar `circleci_ip_ranges` em projetos com Docker

Guia para habilitar IP fixo do CircleCI em projetos que fazem `docker build` no pipeline.

## Por que essa arquitetura

`circleci_ip_ranges: true` é **incompatível** com `setup_remote_docker` no mesmo job (o CircleCI rejeita o job). A solução é dividir o pipeline em dois tipos de job:

| Tipo de job | `circleci_ip_ranges` | `setup_remote_docker` | Função |
|---|---|---|---|
| `secure-checkout` | ✅ true | ❌ não usa | Clona o código de IP fixo, persiste workspace |
| `build` / `test` | ❌ não usa | ✅ usa | Recebe workspace, roda docker build/push |
| `deploy` | ✅ true | ❌ não usa | Acessa ArgoCD/infra interna de IP fixo |

O clone e o deploy saem de IP fixo (atendem política de segurança). O build roda em sandbox isolado de Remote Docker — não toca infra interna, só fala com ECR via IAM.

## Pré-requisitos

Antes de aplicar, confirme que **no seu projeto**:

- ✅ Os builds só puxam imagens do ECR (autenticado por IAM, não IP)
- ✅ Os testes não acessam recursos internos via rede durante a execução

## Passo a passo

### 1. Adicionar o job `secure-checkout`

Cole no topo da seção `jobs:` do seu `.circleci/config.yml`:

```yaml
jobs:
  secure-checkout:
    circleci_ip_ranges: true
    docker:
      - image: <SUA_IMAGEM_BASE>
    working_directory: ~/<NOME_DO_PROJETO>/
    steps:
      - checkout
      - persist_to_workspace:
          root: ~/<NOME_DO_PROJETO>
          paths:
            - .
```

Substitua:
- `<SUA_IMAGEM_BASE>`: a mesma imagem usada pelos outros jobs
- `<NOME_DO_PROJETO>`: o working_directory que seu projeto usa

### 2. Nos jobs com `setup_remote_docker`: trocar `checkout` por `attach_workspace`

Em **cada job que faz `docker build`**:

**Antes:**
```yaml
build:
  steps:
    - checkout                  # ❌ remove
    - setup_remote_docker:
        docker_layer_caching: true
    - run: docker build ...
```

**Depois:**
```yaml
build:
  steps:
    - attach_workspace:         # ✅ substitui o checkout
        at: ~/<NOME_DO_PROJETO>
    - setup_remote_docker:
        docker_layer_caching: true
    - run: docker build ...
```

⚠️ **Não adicione** `circleci_ip_ranges: true` nesses jobs — o build vai ser rejeitado pelo CircleCI.

### 3. Nos jobs de deploy (sem Remote Docker): adicionar `circleci_ip_ranges`

Em jobs que **não usam Remote Docker** (ex: deploy via ArgoCD, kubectl, scripts de infra):

```yaml
deploy:
  circleci_ip_ranges: true      # ✅ adiciona
  steps:
    - checkout:
        depth: 1
    - run: ./deploy.sh
```

Esses jobs podem fazer seu próprio `checkout` normalmente — não precisam do workspace.

### 4. Atualizar o workflow

Adicionar `secure-checkout` como **primeiro job** e dependência dos jobs de build/test:

```yaml
workflows:
  meu-workflow:
    jobs:
      - secure-checkout:
          context: <SEU_CONTEXT>

      - test:
          requires:
            - secure-checkout    # ✅ depende do checkout seguro
          # ...

      - build:
          requires:
            - test               # workspace persiste pelo workflow inteiro
          # ...

      - deploy:
          requires:
            - build
          # ...
```

⚠️ **Regra do `requires`**: todo job que usa `attach_workspace` precisa, **direta ou indiretamente**, depender do `secure-checkout`. Se você adicionar um job novo (ex: `scan-audit`, `lint`, etc.) sem `requires: secure-checkout`, ele pode rodar antes do workspace existir e quebrar em runtime — esse erro **não é pego pela validação de schema**.

Exemplos válidos:
- `build` com `requires: [secure-checkout]` ✅
- `deploy` com `requires: [build]` ✅ (transitivamente depende de secure-checkout)
- `scan-audit` com `requires: [secure-checkout]` ✅ (paralelo aos testes, mas depende do checkout)

## Validação após aplicar

1. Validar YAML localmente:
   ```bash
   python3 -c "import yaml; yaml.safe_load(open('.circleci/config.yml'))"
   ```
2. Validar o schema do CircleCI (pega erros de estrutura que o YAML puro não pega):
   ```bash
   circleci config validate .circleci/config.yml
   ```
3. Push num branch de teste (não em `master`/`main`)
4. Abrir o primeiro run do `secure-checkout` no CircleCI e confirmar que o IP de saída está no pool documentado em https://circleci.com/docs/ip-ranges/
5. Confirmar que os jobs de `build`/`test` rodam sem erro de Remote Docker

## Erros comuns

### `extraneous key [at] is not permitted`

Indentação errada no `attach_workspace`. O `at:` precisa ser **filho** da chave `attach_workspace:`, não irmão.

**Errado** (faltam 2 espaços antes de `at:`):
```yaml
steps:
  - attach_workspace:
    at: ~/projeto       # ❌ no mesmo nível de attach_workspace
```

**Certo:**
```yaml
steps:
  - attach_workspace:
      at: ~/projeto     # ✅ filho de attach_workspace
```

### `This job was rejected because the IP ranges feature does not support Remote Docker`

Você adicionou `circleci_ip_ranges: true` em um job que tem `setup_remote_docker`. Remova o `circleci_ip_ranges` desse job — ele só pertence aos jobs `secure-checkout` e `deploy-*` (que não usam Remote Docker).

### `Cannot find a valid workspace` ou job com diretório vazio

O job está usando `attach_workspace` mas não tem `requires: secure-checkout` (direto ou transitivo) no workflow. Adicione a dependência.

### Job com `attach_workspace` rodando antes da hora

Mesma causa do anterior — falta o `requires`. O CircleCI não infere isso automaticamente do uso do `attach_workspace`.

## Justificativa para o time de segurança

> Toda chamada de rede da plataforma CircleCI sai de IPs conhecidos (clone do código + deploys). A VM de Remote Docker é um sandbox efêmero que apenas autentica no ECR via IAM e roda containers locais — não estabelece conexão com infra interna da empresa, então o IP de saída dela não representa superfície de risco.

## Custo / impacto

- **+1 job** no início do workflow (~30s a 1min de billing)
- **Workspace** trafega o repo entre jobs (transparente, alguns MB)
- Nenhuma mudança em segredos/contexts existentes


# PRs de exemplo: 
https://bitbucket.org/sistema_vert/dev-verttec/pull-requests/8822
https://bitbucket.org/sistema_vert/dev-verttec/pull-requests/8826
