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

      rendererMain.setSize(window.innerWidth, window.innerHeight);
      rendererSecondary.setSize(window.innerWidth / 5, window.innerHeight / 5);
      document.getElementById("scene").appendChild(rendererMain.domElement);
      document
         .getElementById("secondary-scene")
         .appendChild(rendererSecondary.domElement);
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

      function onWindowResize() {
         cameraMain.aspect = window.innerWidth / window.innerHeight;
         cameraSecondary.aspect = window.innerWidth / window.innerHeight;
         cameraMain.updateProjectionMatrix();
         cameraSecondary.updateProjectionMatrix();

         rendererMain.setSize(window.innerWidth, window.innerHeight);
         rendererMain.render(sceneReal, cameraMain);

         rendererSecondary.setSize(
            window.innerWidth / 5,
            window.innerHeight / 5
         );
         rendererSecondary.render(sceneSimu, cameraSecondary);
      }

      window.addEventListener("resize", onWindowResize, false);
      App.scene.data = { scene: sceneReal, camera: cameraMain };
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
      const material = new THREE.MeshLambertMaterial({ color: 0xffff00 });
      const yOffset = 0.5;

      // Create all objects
      App.state.drones = locations.map((l) => {
         let drone = new THREE.Mesh(geometry, material);
         drone.position.x = l.X;
         drone.position.y = l.Y + yOffset;
         drone.position.z = l.Z;
         App.scene.data.scene.add(drone);
         return drone;
      });
      App.state.locations = locations;
   },
   updateDrones: (drones, locations) => {
      App.state.drones = drones;
      App.state.locations = locations;
   },
   updateDrone: (droneId, location) => {
      App.state.drones[droneId].position.x = location.X;
      App.state.drones[droneId].position.y = location.Y;
      App.state.drones[droneId].position.z = location.Z;
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

      document.getElementById("pattern-up").onclick = function () {
         send({ Targets: moveUp(App.state.locations, 5) });
         App.ui.updateStatus(false);
      };

      // Swap
      document.getElementById("swap").onclick = () => {
         App.scene.swap();
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
