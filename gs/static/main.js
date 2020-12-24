import * as THREE from "https://unpkg.com/three@0.123/build/three.module.js";
import { OrbitControls } from "https://unpkg.com/three@0.123/examples/jsm/controls/OrbitControls.js";
// import { moveUp } from "./patterns";

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
   return { scene, camera };
};

// Create drones
const createDrones = ({ scene }, dronesLocation) => {
   const geometry = new THREE.ConeGeometry(0.5, 1, 32);
   const material = new THREE.MeshLambertMaterial({ color: 0xffff00 });
   const yOffset = 0.5;

   // Create all objects
   const drones = dronesLocation.map((l) => {
      let drone = new THREE.Mesh(geometry, material);
      drone.position.x = l.x;
      drone.position.y = l.y + yOffset;
      drone.position.z = l.z;
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

const handleMessage = (sceneData, message) => {
   if (message.Identifier != null) {
      document.getElementById("identifier").innerHTML = message.Identifier;
   }

   if (message.Drones != null && Array.isArray(message.Drones)) {
      document.getElementById("nbDrone").innerHTML = message.Drones.length;
      createDrones(sceneData, message.Drones);
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

   const sceneData = initScene();

   conn.onmessage = function (evt) {
      const message = JSON.parse(evt.data);
      handleMessage(sceneData, message);
      console.log(message);
   };
} else {
   var item = document.createElement("div");
   item.innerHTML = "<b>Your browser does not support WebSockets.</b>";
   console.log(item);
}
