package impl

import (
	"encoding/json"
	"github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/impl/cl_client_network"
	"github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/impl/cl_client_network/teku"
	"github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/impl/el_client_network"
	"github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/impl/el_client_network/geth"
	"github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/impl/forkmon"
	"github.com/kurtosis-tech/eth2-merge-kurtosis-module/kurtosis-module/impl/prelaunch_data_generator"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"path"
	"text/template"
	"time"

	// "path"
	// "text/template"
)

const (
	networkId = "3151908"

	// The number of validator keys that will be preregistered inside the CL genesis file when it's created
	numValidatorsToPreregister = 100

	// NOTE: We saw issues 1 Geth node & 3 Teku nodes was causing problems, and the Teku folks
	//  let us know that generally each CL node should be paired with 1 EL node
	// https://discord.com/channels/697535391594446898/697539289042649190/922266717667856424
	numElAndClPairs = 1

	// ----------------------------------- Genesis Config Constants -----------------------------------------
	// We COULD drop this, but it won't represent mainnet
	secondsPerSlot = uint32(12)
	altairForkEpoch = uint64(1)  // Set per Parithosh's recommendation
	mergeForkEpoch = uint64(2)   // Set per Parithosh's recommendation
	// TODO Should be set to roughly one hour (??) so that this is reached AFTER the CL gets the merge fork version (per Parithosh)
	totalTerminalDifficulty  = uint64(60000000)

	// This is the mnemonic that will be used to generate validator keys which will be preregistered in the CL genesis.ssz that we create
	// This is the same mnemonic that should be used to generate the validator keys that we'll load into our CL nodes when we run them
	preregisteredValidatorKeysMnemonic = "giant issue aisle success illegal bike spike question tent bar rely arctic volcano long crawl hungry vocal artwork sniff fantasy very lucky have athlete"
	// --------------------------------- End Genesis Config Constants ----------------------------------------

	// ----------------------------------- Static File Constants -----------------------------------------
	staticFilesDirpath                    = "/static-files"

	// Geth + CL genesis generation
	genesisGenerationConfigDirpath = staticFilesDirpath + "/genesis-generation-config"
	gethGenesisGenerationConfigYmlTemplateFilepath = genesisGenerationConfigDirpath + "/el/genesis-config.yaml.tmpl"
	clGenesisGenerationConfigYmlTemplateFilepath = genesisGenerationConfigDirpath + "/cl/config.yaml.tmpl"
	clGenesisGenerationMnemonicsYmlTemplateFilepath = genesisGenerationConfigDirpath + "/cl/mnemonics.yaml.tmpl"

	// Nethermind
	nethermindGenesisJsonTemplateFilepath = staticFilesDirpath + "/nethermind-genesis.json.tmpl"

	// Forkmon config
	forkmonConfigTemplateFilepath = staticFilesDirpath + "/forkmon-config/config.toml.tmpl"
	// --------------------------------- End Static File Constants ----------------------------------------

	responseJsonLinePrefixStr = ""
	responseJsonLineIndentStr = "  "

	// TODO uncomment these when the module can either start a private network OR connect to an existing devnet
	// mergeDevnet3NetworkId = "1337602"
	// mergeDevnet3ClClientBootnodeEnr = "enr:-Iq4QKuNB_wHmWon7hv5HntHiSsyE1a6cUTK1aT7xDSU_hNTLW3R4mowUboCsqYoh1kN9v3ZoSu_WuvW9Aw0tQ0Dxv6GAXxQ7Nv5gmlkgnY0gmlwhLKAlv6Jc2VjcDI1NmsxoQK6S-Cii_KmfFdUJL2TANL3ksaKUnNXvTCv1tLwXs0QgIN1ZHCCIyk"
)
/*
var mergeDevnet3BootnodeEnodes = []string{
	"enode://6b457d42e6301acfae11dc785b43346e195ad0974b394922b842adea5aeb4c55b02410607ba21e4a03ba53e7656091e2f990034ce3f8bad4d0cca1c6398bdbb8@137.184.55.117:30303",
	"enode://588ef56694223ce3212d7c56e5b6f3e8ba46a9c29522fdc6fef15657f505a7314b9bd32f2d53c4564bc6b9259c3d5c79fc96257eff9cd489004c4d9cbb3c0707@137.184.203.157:30303",
	"enode://46b2ecd18c24463413b7328e9a59c72d955874ad5ddb9cd9659d322bedd2758a6cefb8378e2309a028bd3cdf2beca0b18c3457f03e772f35d0cd06c37ce75eee@137.184.213.208:30303",
}
 */

type ExecuteResponse struct {
	ForkmonPublicURL string	`json:"forkmonUrl"`
}

type ExampleExecutableKurtosisModule struct {
}

func NewExampleExecutableKurtosisModule() *ExampleExecutableKurtosisModule {
	return &ExampleExecutableKurtosisModule{}
}

func (e ExampleExecutableKurtosisModule) Execute(enclaveCtx *enclaves.EnclaveContext, serializedParams string) (serializedResult string, resultError error) {
	logrus.Info("Generating prelaunch data...")
	genesisUnixTimestamp := time.Now().Unix()
	gethGenesisConfigTemplate, err := parseTemplate(gethGenesisGenerationConfigYmlTemplateFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the Geth genesis generation config YAML template")
	}
	clGenesisConfigTemplate, err := parseTemplate(clGenesisGenerationConfigYmlTemplateFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the CL genesis generation config YAML template")
	}
	clGenesisMnemonicsYmlTemplate, err := parseTemplate(clGenesisGenerationMnemonicsYmlTemplateFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the CL mnemonics YAML template")
	}
	prelaunchData, err := prelaunch_data_generator.GeneratePrelaunchData(
		enclaveCtx,
		gethGenesisConfigTemplate,
		clGenesisConfigTemplate,
		clGenesisMnemonicsYmlTemplate,
		preregisteredValidatorKeysMnemonic,
		numValidatorsToPreregister,
		numElAndClPairs,
		genesisUnixTimestamp,
		networkId,
		secondsPerSlot,
		altairForkEpoch,
		mergeForkEpoch,
		totalTerminalDifficulty,
		preregisteredValidatorKeysMnemonic,
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred launching the Ethereum genesis generator Service")
	}
	logrus.Info("Successfully generated prelaunch data")

	// TODO Nethermind template-filling here
	/*
	tmpl, err := template.New(templateFilename).ParseFiles(templateFilepath)
	template.New(
		// For some reason, the template name has to match the basename of the file:
		//  https://stackoverflow.com/questions/49043292/error-template-is-an-incomplete-or-empty-template
		path.Base(nethermindGenesisJsonTemplateFilepath),
	).Parse(
		gethGenesisJsonFilepath,
	)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing the Nethermind genesis JSON template file '%v'", nethermindGenesisJsonTemplateFilepath)
	}
	 */

	logrus.Info("Launching a network of EL clients...")
	gethClientLauncher := geth.NewGethELClientLauncher(prelaunchData.GethELGenesisJsonFilepathOnModuleContainer)
	elNetwork := el_client_network.NewExecutionLayerNetwork(
		enclaveCtx,
		networkId,
		gethClientLauncher,
	)

	allElClientContexts := []*el_client_network.ExecutionLayerClientContext{}
	for i := 0; i < numElAndClPairs; i++ {
		elClientCtx, err := elNetwork.AddNode()
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred adding EL client node %v", i)
		}
		allElClientContexts = append(allElClientContexts, elClientCtx)
	}
	logrus.Info("Successfully launched a network of EL clients")

	logrus.Info("Launching a network of CL clients...")
	clGenesisPaths := prelaunchData.CLGenesisPaths
	// clClientLauncher := lighthouse.NewLighthouseCLClientLauncher(clGenesisPaths.GetParentDirpath())
	clClientLauncher := teku.NewTekuCLClientLauncher(
		clGenesisPaths.GetConfigYMLFilepath(),
		clGenesisPaths.GetGenesisSSZFilepath(),
	)
	keystoresGenerationResult := prelaunchData.KeystoresGenerationResult
	clNetwork := cl_client_network.NewConsensusLayerNetwork(
		enclaveCtx,
		allElClientContexts,
		clClientLauncher,
		keystoresGenerationResult.PerNodeKeystoreDirpaths,
	)

	allClClientContexts := []*cl_client_network.ConsensusLayerClientContext{}
	for i := 0; i < numElAndClPairs; i++ {
		clClientCtx, err := clNetwork.AddNode()
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred adding CL client node %v", i)
		}
		allClClientContexts = append(allClClientContexts, clClientCtx)
	}
	logrus.Info("Successfully launched a network of CL clients")

	logrus.Info("Launching forkmon...")
	forkmonConfigTemplate, err := parseTemplate(forkmonConfigTemplateFilepath)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred parsing forkmon config template file '%v'", forkmonConfigTemplateFilepath)
	}
	forkmonPublicUrl, err := forkmon.LaunchForkmon(
		enclaveCtx,
		forkmonConfigTemplate,
		allClClientContexts,
		genesisUnixTimestamp,
		secondsPerSlot,
	)
	logrus.Info("Successfully launched forkmon")

	responseObj := &ExecuteResponse{
		ForkmonPublicURL: forkmonPublicUrl,
	}
	responseStr, err := json.MarshalIndent(responseObj, responseJsonLinePrefixStr, responseJsonLineIndentStr)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serializing the following response object to JSON for returning: %+v", responseObj)
	}

	return string(responseStr), nil
}

func parseTemplate(filepath string) (*template.Template, error) {
	tmpl, err := template.New(
		// For some reason, the template name has to match the basename of the file:
		//  https://stackoverflow.com/questions/49043292/error-template-is-an-incomplete-or-empty-template
		path.Base(filepath),
	).ParseFiles(
		filepath,
	)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred parsing template file '%v'", filepath)
	}
	return tmpl, nil
}