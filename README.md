# HackinTN Bot But in Go
 Etapes de bases pour crée une commande

## Créer une commande simple
 ### Sa fonction
La fonction de la commande doit respecter cette structure : `cmdContext` en entrée et pas d'argument de sorti.
Le context permet d'accéder à la méthode `reply` qui permet de répondre à n'importe quel type de commande (classique, composant ou Slash Commands)
D'autres arguments peuvent être spécifié `FollowUp`, `Edit`, `Delete`, `ChannelID`, `ID` pour mieux répondre etc..
```go
func ping(ctx *cmdContext) {
	ctx.reply(replyParams{
		Content:   "Pong!",
		Ephemeral: true,
	})
}
```
 ### Charger la commande
D'autres argument peuvent être spécifié pour créer des commandes avec des arguments par exemple. Il est aussi possible d'enlever certains arguments pour ne pas générer de Commande Slash associé. (`Menu`)

 ```go
 ping := &Command{
	Name:        "ping",
	Description: "Recevoir Pong!",
	Aliases:     cmdAlias{"p"},
	Menu:        GeneralMenu,
	Call:        ping,
 }

addCmd(ping)
 ```
