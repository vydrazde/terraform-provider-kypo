package main

import "terraform-provider-kypo/internal/KYPOClient"

func main() {
	client := KYPOClient.NewClient("https://images.crp.kypo.muni.cz/kypo-sandbox-service/api/v1",
		"***")

	_, _ = client.CreateDefinition("git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git", "master")

}
