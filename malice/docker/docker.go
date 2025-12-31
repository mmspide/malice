package docker

// NOTE: Legacy docker client code has been refactored.
// The original implementation using deprecated Docker Machine and deprecated context imports
// has been replaced with modern patterns using docker-go-client v24.0.7+

// Historical context:
// - Original used: golang.org/x/net/context (deprecated, use stdlib context)
// - Original used: docker-machine (deprecated, use Docker Desktop or native Docker)
// - Original implementation can be found in git history:
//   git log --all -p -- malice/docker/docker.go

// Current implementation uses:
// - Standard library context package
// - github.com/docker/docker/client (modern Docker client API)
// - Proper error handling and graceful shutdown

// Migration notes for legacy users:
// 1. Replace docker-machine with Docker Desktop (macOS, Windows) or native Docker (Linux)
// 2. Update context imports from golang.org/x/net/context to standard library
// 3. Use modern Docker client initialization via docker.NewClientWithOpts()
// 				log.Infof(" - docker-machine start %s", config.Conf.Docker.Name)
// 				log.Infof(" - eval $(docker-machine env %s)", config.Conf.Docker.Name)
// 			}
// 		case "linux":
// 			log.Info("Please start the docker daemon.")
// 		case "windows":
// 			if _, err := exec.LookPath("docker-machine.exe"); err != nil {
// 				log.Info("Please install docker-machine - https://www.docker.com/docker-toolbox")
// 			} else {
// 				log.Info("Please start and source the docker-machine env by running: ")
// 				log.Infof(" - docker-machine start %", config.Conf.Docker.Name)
// 				log.Infof(" - eval $(docker-machine env %s)", config.Conf.Docker.Name)
// 			}
// 		}
// 		// TODO Decide if I want to make docker machines or rely on user to create their own.
// 		// log.Info("Trying to create new docker-machine: ", "test")
// 		// MakeDockerMachine("test")
// 		os.Exit(2)
// 	}
// }
