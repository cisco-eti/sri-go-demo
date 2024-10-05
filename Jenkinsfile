@Library(['srePipeline']) _

// --------------------------------------------
// Refer to Pipeline docs for options used in mysettings
// https://wwwin-github.cisco.com/pages/eti/sre-pipeline-library
// --------------------------------------------

def pipelinesettings = [
  deploy: [
    [name: "sre-go-helloworld" ]                              // Containers to publish
  ],

  gitDefaultBranch: "main",
  prepare: 1,                                                 // GIT Clone
  unittest: 1,                                                // Unit-test
  build: 1,                                                   // Build container
  executeCC: 1,                                               // Generate Code Coverage report
  lint: 1,                                                    // GO Lint
  sonarQube: [                                                // Corona paramters
    propertiesFile: "./build/sonar-project.properties"        // SonarQube scan
  ],
  panoptica: [                                                // Panoptica Scan Parameters
    productName: "securecn",                                  // The Product/Venture name
    imageName: "sre-go-helloworld",                           // The image name + tag to scan
    args: "--highest-severity-allowed HIGH --ignore-no-fix",  // Arguments passed to the Panoptica CLI
  ],
  publishContainer: 1,                                        // Publish container
  registryOrgName: "eti-sre/",                                // Container image registry org name
  ecr: 1,                                                     // Publish container to Private ECR
  ciscoContainer: 1,                                          // Publish container to containers.cisco.com
  dockerHub: 1,                                               // Publish container to dockerhub.cisco.com
  pushPublicRegistryOnTag: 1,                                 // Publish container to Public ECR on tag
  // forceCorona: 1,                                          // Force Corona Scan on any branch
  corona: [                                                   // Corona paramters
    imageName: "sre-go-helloworld",                           // Corona Image Name
    releaseID: "73243",                                       // Corona Release ID
    productID: "6726",                                        // Corona Project ID
    csdlID: "84720",                                          // Corona CSDL ID
    securityContact: "sraradhy@cisco.com",                    // Corona Security Contact
    engineeringContact: "sraradhy@cisco.com",                 // Corona Engineering Contact
    imageAdmins: "sraradhy,jegarnie",                         // Corona Image Admins
  ],
  // forceBlackduck: 1,                                       // Force Blackduck Scan on any branch
  blackduck: [
    email: "eti-sre-admins@cisco.com",
  ],                                                          // Blackduck Open Source Scan

  chart: "deploy/charts",                                     // Use deployment multi-helm chart folder
  publishHelm: 1,                                             // HELM CREATE
  useMultipleHelm:1,                                          // Publish Multiple Charts
  chartMuseum: 1,                                             // Publish Helm chart to ChartMuseum
  artifactory: 1,                                             // Use Artifactory creds

  stricterCCThreshold: 90.0,                                  // Fail builds for Code Coverage below 90%
  cdPromotionJobPath: "../../deploy/dev/sre-go-helloworld-cd",
]

srePipeline( pipelinesettings )

