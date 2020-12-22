// import { OrbitControls } from './jsm/controls/OrbitControls.js';

const scene = new THREE.Scene();
const camera = new THREE.PerspectiveCamera(
  75,
  window.innerWidth / window.innerHeight,
  0.1,
  1000
);

const renderer = new THREE.WebGLRenderer();
renderer.setSize(window.innerWidth, window.innerHeight);
document.getElementById('scene').appendChild(renderer.domElement);
// document.body.prepend(renderer.domElement);

const geometry = new THREE.ConeGeometry(0.5, 1, 32);
const material = new THREE.MeshLambertMaterial({ color: 0xffff00 });

// controls

// const controls = new OrbitControls( camera, renderer.domElement );
// controls.minDistance = 20;
// controls.maxDistance = 50;
// controls.maxPolarAngle = Math.PI / 2;

// helper

scene.add(new THREE.AxesHelper(20));

// light

scene.add(new THREE.AmbientLight(0x222222));
const light = new THREE.PointLight(0xffffff, 1);
light.position.set(50, 50, 50);
scene.add(light);

// const cone = new THREE.Mesh( geometry, material );
// scene.add( cone );

// TODO Create all objects
const drones = [];
const spacing = 3;
const nbDrones = 9;
const yOffset = 0.5;

for (let i = 0; i < nbDrones; ++i) {
  let drone = new THREE.Mesh(geometry, material);
  drone.position.x = (i % 3) * spacing;
  drone.position.y = yOffset;
  drone.position.z = Math.floor(i / 3) * spacing;
  drones.push(drone);
  scene.add(drone);
}

camera.position.x = 10;
camera.position.y = 10;
camera.position.z = 10;
camera.lookAt(new THREE.Vector3(0, 0, 0));
renderer.render(scene, camera);

// const animate = function () {
//   requestAnimationFrame(animate);

//   cube.rotation.x += 0.01;
//   cube.rotation.y += 0.01;

//   renderer.render(scene, camera);
// };

// animate();

function onWindowResize() {
  camera.aspect = window.innerWidth / window.innerHeight;
  camera.updateProjectionMatrix();

  renderer.setSize( window.innerWidth, window.innerHeight );
  renderer.render(scene, camera);
}

window.addEventListener( 'resize', onWindowResize, false );
