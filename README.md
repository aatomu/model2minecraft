# model2minecraft

This program to convert `.obj`,`.png/.jpg`,`.mp4` to arbitrary `.mcfunction` \
implemented only in the standard library / pure golang

## configuration

Example configuration: to see main.go

### output configuration

|       key        | type                                            | example | description                                |
| :--------------: | :---------------------------------------------- | :------ | :----------------------------------------- |
|    sourceType    | Source(enum: `Object`/`Image`/`Video`)          | Object  | \*ffmpeg is required to use `Video`        |
| maxCommandChain  | int                                             | 50000   | minecraft default maxCommandChain is 65535 |
|  colorDepthBit   | int(range: `1..8`)                              | 4       |                                            |
| commandGenerator | Command(func(arg CommandArgument) (cmd string)) |         |                                            |
| enableBlockCount | bool                                            | true    | when true,output used block count          |

### sourceType=Object configuration

|        key        | type   | example         | description                                        |
| :---------------: | :----- | :-------------- | :------------------------------------------------- |
|  objectDirectory  | string | ./3d            | resource directory of object files                 |
|  objectFilename   | string | HatsuneMiku.obj |                                                    |
|    objectScale    | Frac   | NewFrac(1/10)   | resizing .obj                                      |
| objectGridSpacing | Frac   | NewFrac(1/1)    | cubic grid spacing                                 |
| isObjectUVYAxisUp | bool   | true            | depends on the creation software                   |
|   parallelLimit   | int    | 500             | the bigger it is, the heavier it gets, but faster. |

### sourceType=Image configuration

|      key      | type   | example       | description |
| :-----------: | :----- | :------------ | :---------- |
| imageFilename | string | ./example.png |             |

### sourceType=Video configuration

|      key       | type   | example       | description                                    |
| :------------: | :----- | :------------ | :--------------------------------------------- |
| videoFilename  | string | ./example.mp4 |                                                |
| videoFrameRate | int    | 20            | video cut fps setting, max 20 fps in Minecraft |
| videoScaleSize | string | 200:-1        | check ffmpeg `-vf` argument                    |

### Minecraft configuration

|        key         | type     | example                                                            | description                  |
| :----------------: | :------- | :----------------------------------------------------------------- | :--------------------------- |
| minecraftDirectory | string   | ./minecraft                                                        | minecraft textures directory |
|  allowedBlockIds   | []string | []string{""}                                                       | \*working regex patterns     |
|  ignoredBlockIds   | []string | []string{"powder$", "sand$", "gravel$", "glass", "spawner", "ice"} | \*working regex patterns     |

Place the file `minecraftDirectory` with the asset files extracted from `version.jar` \
Example: `version.jar/assets/minecraft/blockstates/stone.json` > `${minecraftDirectory}/minecraft/blockstates/stone.json`
