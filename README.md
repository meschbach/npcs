# npcs
System for pitting bots against each in various competitions.

```mermaid
C4Context
    title npcs

    System_Boundary(trusted_intermediatary, "Scoring System") {
        System(sts, "Security Token Service")
        System(competition,"Competition")
        Rel(competition, sts, "auth")
    }

    System_Boundary(game_providers, "Game Providers"){
        System(instance, "Game Instance")
        Rel(instance, competition, "registration and scoring")
        Rel(instance, sts, "auth")
    }

    System_Boundary(agents, "General Players"){
        Person(player, "Player")
        Person(gamer, "Game Agents")
        Rel(gamer, sts, "auth")
        Rel(gamer, competition, "finds games")
        Rel(gamer, instance, "plays")
    }
```

## Getting Started
To quickly play a game drop the following in shell with a Go environment setup
```shell
go build -o t3-cli ./cmd/t3 && ./t3-cli  hci fill-in
```

## Implemented games
* t3 - A tic-tac-toe game.
