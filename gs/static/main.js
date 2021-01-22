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

let initScene = () => {
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
   return { scene, camera, drones: [], dronesLocation: [] };
};

// Create drones
const createDrones = ({ scene }, dronesLocation) => {
   const geometry = new THREE.ConeGeometry(0.5, 1, 32);
   const material = new THREE.MeshLambertMaterial({ color: 0xffff00 });
   const yOffset = 0.5;

   // Create all objects
   const drones = dronesLocation.map((l) => {
      let drone = new THREE.Mesh(geometry, material);
      drone.position.x = l.X;
      drone.position.y = l.Y + yOffset;
      drone.position.z = l.Z;
      scene.add(drone);
      return drone;
   });

   return drones;
};

const fakeDronesLocation = () => {
   const drones = [];
   const spacing = 3;
   const nbDrones = 9;

   for (let i = 0; i < nbDrones; ++i) {
      let drone = {
         x: (i % 3) * spacing,
         y: 0,
         z: Math.floor(i / 3) * spacing,
      };
      drones.push(drone);
   }
   return drones;
};

const App = () => ({});
App.state = {
   drones: [],
   dronesLocation: [],
   updateDrones: (drones, locations) => {
      App.state.drones = drones;
      App.state.dronesLocation = locations;
   },
};

const handleMessage = (sceneData, message) => {
   if (message.Identifier != null) {
      document.getElementById("identifier").innerHTML = message.Identifier;
   }

   if (message.Drones != null && Array.isArray(message.Drones)) {
      document.getElementById("nbDrone").innerHTML = message.Drones.length;
      const drones = createDrones(sceneData, message.Drones);
      App.state.updateDrones(drones, message.Drones);
   }
};

const uiHandlerSetup = (send) => {
   const initial = document.getElementById("pattern-initial");
   const up = document.getElementById("pattern-up");
   const spherical = document.getElementById("pattern-spherical");

   up.onclick = function () {
      console.log(App.state);
      console.log(moveUp(App.state.dronesLocation, 5));
      send({ Targets: moveUp(App.state.dronesLocation, 5) });
   };
};

// WebSocket
if (window["WebSocket"]) {
   let conn = new WebSocket("ws://" + document.location.host + "/ws");
   conn.onclose = function (evt) {
      var item = document.createElement("div");
      item.innerHTML = "<b>Connection closed.</b>";
      console.log(item);
   };

   const sceneData = initScene();

   conn.onmessage = function (evt) {
      console.log(evt.data);
      const message = JSON.parse(evt.data);
      handleMessage(sceneData, message);
      console.log(message);
   };

   uiHandlerSetup((data) => {
      console.log("Send data", data);
      conn.send(JSON.stringify(data));
   });
} else {
   var item = document.createElement("div");
   item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
   console.log(item);
}
