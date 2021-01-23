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
   data: {},
   init: () => {
      const scene = new THREE.Scene();
      const camera = new THREE.PerspectiveCamera(
         75,
         window.innerWidth / window.innerHeight,
         0.1,
         1000
      );

      const renderer = new THREE.WebGLRenderer();
      renderer.setSize(window.innerWidth, window.innerHeight);
      document.getElementById("scene").appendChild(renderer.domElement);
      // document.body.prepend(renderer.domElement);

      // controls
      const controls = new OrbitControls(camera, renderer.domElement);
      controls.minDistance = 20;
      controls.maxDistance = 50;
      controls.maxPolarAngle = Math.PI / 2;
      controls.update();

      // helper
      scene.add(new THREE.AxesHelper(20));

      // light
      scene.add(new THREE.AmbientLight(0x222222));
      const light = new THREE.PointLight(0xffffff, 1);
      light.position.set(50, 50, 50);
      scene.add(light);

      camera.position.set(10, 10, 10);
      camera.lookAt(new THREE.Vector3(0, 0, 0));
      controls.update();

      // Render
      const animate = function () {
         requestAnimationFrame(animate);

         controls.update();

         renderer.render(scene, camera);
      };
      animate();

      function onWindowResize() {
         camera.aspect = window.innerWidth / window.innerHeight;
         camera.updateProjectionMatrix();

         renderer.setSize(window.innerWidth, window.innerHeight);
         renderer.render(scene, camera);
      }

      window.addEventListener("resize", onWindowResize, false);
      App.scene.data = { scene, camera };
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
      // App.scene.data.scene.
   },
};

App.ui = {
   init: (send) => {
      App.ui.updateStatus(true);
      App.ui.updateNbDrones(0);
      App.ui.updateIdentifier("Unknown");

      const initial = document.getElementById("pattern-initial");
      const up = document.getElementById("pattern-up");
      const spherical = document.getElementById("pattern-spherical");

      up.onclick = function () {
         send({ Targets: moveUp(App.state.locations, 5) });
         App.ui.updateStatus(false);
      };
      //TODO: Implement Spherical and initial
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
      console.log(item);
   };

   App.scene.init();
   App.ui.init((data) => {
      console.log("Send data", data);
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
