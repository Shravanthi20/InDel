export default function Overview() {
  return (
    <div className="p-6">
      <h1 className="text-3xl font-bold">Platform Overview</h1>
      <div className="grid grid-cols-2 gap-6 mt-6">
        <div className="bg-white p-6 rounded shadow">
          <h2 className="text-lg font-semibold">Active Workers</h2>
          <p className="text-4xl font-bold mt-2">145</p>
        </div>
        <div className="bg-white p-6 rounded shadow">
          <h2 className="text-lg font-semibold">Live Order Volume</h2>
          <p className="text-4xl font-bold mt-2">2,340</p>
        </div>
      </div>
    </div>
  )
}
