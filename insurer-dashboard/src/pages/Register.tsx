import { useEffect, useState } from 'react';
import { getZones } from '../api/insurer';
import { Zone } from '../types';

export default function Register() {
  const [zones, setZones] = useState<Zone[]>([]);
  const [selectedZone, setSelectedZone] = useState<number | null>(null);

  useEffect(() => {
    getZones().then(res => {
      setZones(res.data.zones);
    });
  }, []);

  return (
    <form className="max-w-md mx-auto mt-8 p-4 border rounded">
      <h2 className="text-xl font-bold mb-4">Register</h2>
      <div className="mb-4">
        <label className="block mb-1">Zone</label>
        <select
          className="w-full border px-2 py-1"
          value={selectedZone ?? ''}
          onChange={e => setSelectedZone(Number(e.target.value))}
        >
          <option value="">Select a zone</option>
          {zones.map(z => (
            <option key={z.zone_id} value={z.zone_id}>
              {z.name} ({z.city})
            </option>
          ))}
        </select>
      </div>
      {/* Add other registration fields here */}
      <button type="submit" className="bg-blue-500 text-white px-4 py-2 rounded">Register</button>
    </form>
  );
}
