-- All entities of those type will be used to determine the screenshot area (in
-- area mode = "entities").
-- From https://wiki.factorio.com/Prototype/Entity
local includes = {
  "character-corpse", -- Prototype/CharacterCorpse
  "corpse", -- Prototype/Corpse
  "rail-remnants", -- Prototype/RailRemnants
  "accumulator", -- Prototype/Accumulator
  "artillery-turret", -- Prototype/ArtilleryTurret
  "beacon", -- Prototype/Beacon
  "boiler", -- Prototype/Boiler
  "burner-generator", -- Prototype/BurnerGenerator
  "character", -- Prototype/Character
  "arithmetic-combinator", -- Prototype/ArithmeticCombinator
  "decider-combinator", -- Prototype/DeciderCombinator
  "constant-combinator", -- Prototype/ConstantCombinator
  "container", -- Prototype/Container
  "logistic-container", -- Prototype/LogisticContainer
  "infinity-container", -- Prototype/InfinityContainer
  "assembling-machine", -- Prototype/AssemblingMachine
  "rocket-silo", -- Prototype/RocketSilo
  "furnace", -- Prototype/Furnace
  "electric-pole", -- Prototype/ElectricPole
  "electric-energy-interface", -- Prototype/ElectricEnergyInterface
  "combat-robot", -- Prototype/CombatRobot
  "construction-robot", -- Prototype/ConstructionRobot
  "logistic-robot", -- Prototype/LogisticRobot
  "gate", -- Prototype/Gate
  "generator", -- Prototype/Generator
  "heat-interface", -- Prototype/HeatInterface
  "heat-pipe", -- Prototype/HeatPipe
  "inserter", -- Prototype/Inserter
  "lab", -- Prototype/Lab
  "lamp", -- Prototype/Lamp
  "land-mine", -- Prototype/LandMine
  "market", -- Prototype/Market
  "mining-drill", -- Prototype/MiningDrill
  "offshore-pump", -- Prototype/OffshorePump
  "pipe", -- Prototype/Pipe
  "infinity-pipe", -- Prototype/InfinityPipe
  "pipe-to-ground", -- Prototype/PipeToGround
  "player-port", -- Prototype/PlayerPort
  "power-switch", -- Prototype/PowerSwitch
  "programmable-speaker", -- Prototype/ProgrammableSpeaker
  "pump", -- Prototype/Pump
  "radar", -- Prototype/Radar
  "curved-rail", -- Prototype/CurvedRail
  "straight-rail", -- Prototype/StraightRail
  "rail-chain-signal", -- Prototype/RailChainSignal
  "rail-signal", -- Prototype/RailSignal
  "reactor", -- Prototype/Reactor
  "roboport", -- Prototype/Roboport
  "solar-panel", -- Prototype/SolarPanel
  "spider-leg", -- Prototype/SpiderLeg
  "storage-tank", -- Prototype/StorageTank
  "train-stop", -- Prototype/TrainStop
  "loader-1x1", -- Prototype/Loader1x1
  "loader", -- Prototype/Loader1x2
  "splitter", -- Prototype/Splitter
  "transport-belt", -- Prototype/TransportBelt
  "underground-belt", -- Prototype/UndergroundBelt
  "ammo-turret", -- Prototype/AmmoTurret
  "electric-turret", -- Prototype/ElectricTurret
  "fluid-turret", -- Prototype/FluidTurret
  "car", -- Prototype/Car
  "artillery-wagon", -- Prototype/ArtilleryWagon
  "cargo-wagon", -- Prototype/CargoWagon
  "fluid-wagon", -- Prototype/FluidWagon
  "locomotive", -- Prototype/Locomotive
  "spider-vehicle", -- Prototype/SpiderVehicle
  "wall", -- Prototype/Wall
  "rocket-silo-rocket", -- Prototype/RocketSiloRocket
  "rocket-silo-rocket-shadow", -- Prototype/RocketSiloRocketShadow
  "speech-bubble", -- Prototype/SpeechBubble
}

local excludes = {
  "arrow", -- Prototype/Arrow
  "artillery-flare", -- Prototype/ArtilleryFlare
  "artillery-projectile", -- Prototype/ArtilleryProjectile
  "beam", -- Prototype/Beam
  "cliff", -- Prototype/Cliff
  "deconstructible-tile-proxy", -- Prototype/DeconstructibleTileProxy
  "entity-ghost", -- Prototype/EntityGhost
  "particle", -- Prototype/EntityParticle
  "leaf-particle", -- Prototype/LeafParticle
  -- "<abstract>", -- Prototype/EntityWithHealth
  -- "<abstract>", -- Prototype/Combinator
  -- "<abstract>", -- Prototype/CraftingMachine
  "unit-spawner", -- Prototype/EnemySpawner
  "fish", -- Prototype/Fish
  -- "<abstract>", -- Prototype/FlyingRobot
  -- "<abstract>", -- Prototype/RobotWithLogisticInterface
  -- "<abstract>", -- Prototype/Rail
  -- "<abstract>", -- Prototype/RailSignalBase
  "simple-entity", -- Prototype/SimpleEntity
  "simple-entity-with-owner", -- Prototype/SimpleEntityWithOwner
  "simple-entity-with-force", -- Prototype/SimpleEntityWithForce
  -- "<abstract>", -- Prototype/TransportBeltConnectable
  "tree", -- Prototype/Tree
  "turret", -- Prototype/Turret [includes biters]
  "unit", -- Prototype/Unit
  -- "<abstract>", -- Prototype/Vehicle
  -- "<abstract>", -- Prototype/RollingStock
  "explosion", -- Prototype/Explosion
  "flame-thrower-explosion", -- Prototype/FlameThrowerExplosion
  "fire", -- Prototype/FireFlame
  "stream", -- Prototype/FluidStream
  "flying-text", -- Prototype/FlyingText
  "highlight-box", -- Prototype/HighlightBoxEntity
  "item-entity", -- Prototype/ItemEntity
  "item-request-proxy", -- Prototype/ItemRequestProxy
  "decorative", -- Prototype/LegacyDecorative
  "particle-source", -- Prototype/ParticleSource
  "projectile", -- Prototype/Projectile
  "resource", -- Prototype/ResourceEntity
  -- "<abstract>", -- Prototype/Smoke
  "smoke", -- Prototype/SimpleSmoke
  "smoke-with-trigger", -- Prototype/SmokeWithTrigger
  "sticker", -- Prototype/Sticker
  "tile-ghost", -- Prototype/TileGhost
}

return {
  includes = includes,
  excludes = excludes,
}