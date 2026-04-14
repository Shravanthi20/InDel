import { useEffect, useState } from 'react';
import { getZonePaths } from '../api/zones';

export default function Register() {
  const [zoneLevel, setZoneLevel] = useState<'a' | 'b' | 'c' | ''>('');
  const [zoneOptions, setZoneOptions] = useState<any[]>([]);
  const [selectedZone, setSelectedZone] = useState('');
  const [zoneCache] = useState<{ [k: string]: any[] }>({});

  useEffect(() => {
    if (!zoneLevel) {
      setZoneOptions([]);
      setSelectedZone('');
      return;
    }
    if (zoneCache[zoneLevel]) {
      setZoneOptions(zoneCache[zoneLevel]);
      setSelectedZone('');
      return;
    }
    getZonePaths(zoneLevel).then(res => {
      const cities = res.data.cities || res.data.zones || [];
      setZoneOptions(cities);
      zoneCache[zoneLevel] = cities;
      setSelectedZone('');
    }).catch(() => setZoneOptions([]));
  }, [zoneLevel, zoneCache]);

  return (
    <form className="max-w-md mx-auto mt-8 p-4 border rounded">
      <h2 className="text-xl font-bold mb-4">Register</h2>
      <div className="mb-4">
        <label className="block mb-1">Zone Level</label>
        <select
          className="w-full border px-2 py-1 mb-2"
          value={zoneLevel}
          onChange={e => setZoneLevel(e.target.value as 'a' | 'b' | 'c' | '')}
        >
          <option value="">Select level</option>
          <option value="a">A</option>
          <option value="b">B</option>
          <option value="c">C</option>
        </select>
        <label className="block mb-1">Zone Name</label>
        <select
          className="w-full border px-2 py-1"
          value={selectedZone}
          onChange={e => setSelectedZone(e.target.value)}
          disabled={!zoneLevel || zoneOptions.length === 0}
        >
          <option value="">{zoneLevel ? 'Select Zone' : 'Select Level First'}</option>
          {zoneOptions.map((z, idx) => (
            <option key={z.city || z.zone_name || idx} value={z.city || z.zone_name}>
              {(z.city || z.zone_name) + (z.state ? ', ' + z.state : '')}
            </option>
          ))}
        </select>
      </div>
      {/* Add other registration fields here */}
      <button type="submit" className="bg-blue-500 text-white px-4 py-2 rounded">Register</button>
    </form>
  );
}
