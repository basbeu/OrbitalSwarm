import * as THREE from "https://unpkg.com/three@0.123/build/three.module.js";
import { OrbitControls } from "https://unpkg.com/three@0.123/examples/jsm/controls/OrbitControls.js";
// import { moveUp } from "./patterns";

const moveUp = (drones, altitude) => {
   return drones.map((d) => ({
      X: d.X,
      Y: d.Y + altitude,
      Z: d.Z,
   }));
};

const App = () => ({});
App.scene = {
   data: {
      swapped: false,
   },
   init: () => {
      const mainSlot = document.getElementById("scene");
      const secondarySlot = document.getElementById("secondary-scene");

      const sceneReal = new THREE.Scene();
      const sceneSimu = new THREE.Scene();
      const cameraMain = new THREE.PerspectiveCamera(
         75,
         window.innerWidth / window.innerHeight,
         0.1,
         1000
      );
      const cameraSecondary = new THREE.PerspectiveCamera(
         75,
         window.innerWidth / window.innerHeight,
         0.1,
         1000
      );

      const rendererMain = new THREE.WebGLRenderer();
      const rendererSecondary = new THREE.WebGLRenderer();
      mainSlot.appendChild(rendererMain.domElement);
      secondarySlot.appendChild(rendererSecondary.domElement);
      rendererMain.setSize(window.innerWidth, window.innerHeight);
      rendererSecondary.setSize(
         secondarySlot.clientWidth,
         secondarySlot.clientWidth * (window.innerHeight / window.innerWidth)
      );

      // document.body.prepend(renderer.domElement);

      // controls
      const controlsMain = new OrbitControls(
         cameraMain,
         rendererMain.domElement
      );
      controlsMain.minDistance = 20;
      controlsMain.maxDistance = 50;
      controlsMain.maxPolarAngle = Math.PI / 2;
      controlsMain.update();

      const controlsSecondary = new OrbitControls(
         cameraSecondary,
         rendererSecondary.domElement
      );
      controlsSecondary.minDistance = 20;
      controlsSecondary.maxDistance = 50;
      controlsSecondary.maxPolarAngle = Math.PI / 2;
      controlsSecondary.update();

      // helper
      sceneReal.add(new THREE.AxesHelper(20));
      sceneSimu.add(new THREE.AxesHelper(20));

      // light
      sceneReal.add(new THREE.AmbientLight(0x222222));
      sceneSimu.add(new THREE.AmbientLight(0x222222));
      const lightReal = new THREE.PointLight(0xffffff, 1);
      const lightSimu = new THREE.PointLight(0xffffff, 1);
      lightReal.position.set(50, 50, 50);
      lightSimu.position.set(50, 50, 50);
      sceneReal.add(lightReal);
      sceneSimu.add(lightSimu);

      cameraMain.position.set(10, 10, 10);
      cameraSecondary.position.set(10, 10, 10);
      cameraMain.lookAt(new THREE.Vector3(0, 0, 0));
      cameraSecondary.lookAt(new THREE.Vector3(0, 0, 0));
      controlsMain.update();
      controlsSecondary.update();

      // Render
      const animate = function () {
         requestAnimationFrame(animate);

         controlsMain.update();
         controlsSecondary.update();

         const swapped = App.scene.data.swapped;
         rendererMain.render(!swapped ? sceneReal : sceneSimu, cameraMain);
         rendererSecondary.render(
            !swapped ? sceneSimu : sceneReal,
            cameraSecondary
         );
      };
      animate();

      const onWindowResize = () => {
         cameraMain.aspect = window.innerWidth / window.innerHeight;
         cameraSecondary.aspect = window.innerWidth / window.innerHeight;
         cameraMain.updateProjectionMatrix();
         cameraSecondary.updateProjectionMatrix();

         rendererMain.setSize(window.innerWidth, window.innerHeight);
         rendererMain.render(sceneReal, cameraMain);

         rendererSecondary.setSize(
            secondarySlot.clientWidth,
            secondarySlot.clientWidth * (window.innerHeight / window.innerWidth)
         );
         rendererSecondary.render(sceneSimu, cameraSecondary);
      };

      window.addEventListener("resize", onWindowResize, false);
      App.scene.data = { sceneReal, sceneSimu };
   },
   swap: () => {
      App.scene.data.swapped = !App.scene.data.swapped;
   },
};

App.state = {
   drones: [],
   locations: [],
   createDrones: (locations) => {
      const geometry = new THREE.ConeGeometry(0.5, 1, 32);
      const materialReal = new THREE.MeshLambertMaterial({ color: 0xffff00 });
      const materialSimu = new THREE.MeshLambertMaterial({ color: 0xff00ff });
      const yOffset = 0.5;

      // Create all objects
      App.state.dronesReal = locations.map((l) => {
         let drone = new THREE.Mesh(geometry, materialReal);
         drone.position.x = l.X;
         drone.position.y = l.Y + yOffset;
         drone.position.z = l.Z;
         App.scene.data.sceneReal.add(drone);
         return drone;
      });
      App.state.dronesSimu = locations.map((l) => {
         let drone = new THREE.Mesh(geometry, materialSimu);
         drone.position.x = l.X;
         drone.position.y = l.Y + yOffset;
         drone.position.z = l.Z;
         App.scene.data.sceneSimu.add(drone);
         return drone;
      });
      App.state.locations = locations;
   },
   startSimulation: (paths) => {
      // 30fps -> 1s par step
      const singleMoveTime = 1000;
      const refreshFrequency = 30;
      const waitingTime = singleMoveTime / refreshFrequency;

      const animateSingleMove = (moves, done) => {
         const singleStep = moves.map((p) => ({
            X: p.X / refreshFrequency,
            Y: p.Y / refreshFrequency,
            Z: p.Z / refreshFrequency,
         }));
         console.log(singleStep);
         const nextMove = (step) => {
            for (let i = 0; i < moves.length; ++i) {
               App.state.dronesSimu[i].position.x += singleStep[i].X;
               App.state.dronesSimu[i].position.y += singleStep[i].Y;
               App.state.dronesSimu[i].position.z += singleStep[i].Z;
            }
            step--;
            if (step > 0) {
               setTimeout(() => nextMove(step), waitingTime);
            } else {
               setTimeout(done, waitingTime);
            }
         };
         nextMove(refreshFrequency);
      };

      const animatePath = (paths) => {
         if (paths[0].length == 0) {
            console.log("Simulation ended");
            return;
         }
         const nextMoves = paths.map((p) => p[0]);
         const remainingMoves = paths.map((p) => p.slice(1));
         setTimeout(
            () =>
               animateSingleMove(nextMoves, () => animatePath(remainingMoves)),
            waitingTime
         );
      };

      console.log("Simulation started");
      animatePath(paths);
   },
   updateDrone: (droneId, location) => {
      App.state.dronesReal[droneId].position.x = location.X;
      App.state.dronesReal[droneId].position.y = location.Y;
      App.state.dronesReal[droneId].position.z = location.Z;
      App.state.locations[droneId] = location;
   },
};


App.ui = {
   init: (send) => {
      App.ui.updateStatus(true);
      App.ui.updateNbDrones(0);
      App.ui.updateIdentifier("Unknown");

      const initial = document.getElementById("pattern-initial");
      const spherical = document.getElementById("pattern-spherical");
      //TODO: Implement Spherical and initial

      document.getElementById("pattern-initial").onclick = () => {
         App.state.startSimulation([
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 0, Z: 1 },
               { X: 1, Y: 0, Z: 0 },
            ], // Drone 1
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 1, Y: 0, Z: 0 },
            ],
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 1, Y: 0, Z: 0 },
            ],
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 1, Y: 0, Z: 0 },
            ],
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
            ],
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 1, Y: 0, Z: 0 },
            ],
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 1, Y: 0, Z: 0 },
            ],
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 1, Y: 0, Z: 0 },
            ],
            [
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 0, Y: 1, Z: 0 },
               { X: 1, Y: 0, Z: 0 },
            ],
         ]);
      };
      document.getElementById("pattern-up").onclick = () => {
         send({ Targets: moveUp(App.state.locations, 5) });
         App.ui.updateStatus(false);
      };

      // Swap
      document.getElementById("swap").onclick = () => {
         App.scene.swap();
         const swapped = App.scene.data.swapped;
         document.getElementById("main-label").textContent = !swapped
            ? "Live data"
            : "Simulation";
         document.getElementById("secondary-label").textContent = !swapped
            ? "Simulation"
            : "Live data";
      };
   },
   updateIdentifier: (identifier) => {
      document.getElementById("identifier").innerHTML = identifier;
   },
   updateNbDrones: (nbDrones) => {
      document.getElementById("nbDrone").innerHTML = nbDrones;
   },
   updateStatus: (ready) => {
      document.getElementById("status").innerHTML = ready
         ? "Waiting for order"
         : "Running ...";
      // TODO: disable buttons
   },
};

const handleMessage = (message) => {
   if (message.Identifier != null) {
      App.ui.updateIdentifier(message.Identifier);
   }

   if (message.Paths != null) {
      App.state.startSimulation(message.Paths);
   }

   if (message.Drones != null && Array.isArray(message.Drones)) {
      App.ui.updateNbDrones(message.Drones.length);
      App.state.createDrones(message.Drones);
   }

   if (message.DroneId != null && message.Location != null) {
      App.state.updateDrone(message.DroneId, message.Location);
   }

   if (message.Ready === true) {
      App.ui.updateStatus(message.Ready);
   }
};

// WebSocket
if (window["WebSocket"]) {
   let conn = new WebSocket("ws://" + document.location.host + "/ws");
   conn.onclose = function (evt) {
      var item = document.createElement("div");
      item.innerHTML = "<b>Connection closed.</b>";
      console.log("Closed connection :'(");
   };

   App.scene.init();
   App.ui.init((data) => {
      console.log("Send data", data);
      console.log("Size : " + JSON.stringify(data).length);
      conn.send(JSON.stringify(data));
   });

   conn.onmessage = function (evt) {
      evt.data.split("\n").forEach((data) => {
         const message = JSON.parse(data);
         handleMessage(message);
      });
   };
} else {
   var item = document.createElement("div");
   item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
   console.log(item);
}
