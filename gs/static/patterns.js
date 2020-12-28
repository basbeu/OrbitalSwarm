const moveUp = (drones, altitude) => {
   return drones.map((d) => ({
      x: d.position.x,
      y: d.position.y + altitude,
      z: d.position.z,
   }));
};

export { moveUp };
