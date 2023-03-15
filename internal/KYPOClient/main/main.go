package main

import "terraform-provider-kypo/internal/KYPOClient"

func main() {
	client, _ := KYPOClient.NewClient("https://images.crp.kypo.muni.cz", "bzhwmbxgyxALbAdMjYOgpolQzkiQHGwWRXxm", "kypo-admin", "***")
	_, _ = client.CreateSandboxDefinition("git@gitlab.ics.muni.cz:muni-kypo-trainings/games/junior-hacker.git", "master")

}
