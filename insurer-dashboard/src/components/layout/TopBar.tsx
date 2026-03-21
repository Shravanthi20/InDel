export default function TopBar() {
  return (
    <div className="bg-white shadow p-6 flex justify-between items-center">
      <h1 className="text-2xl font-bold">InDel Insurer Portal</h1>
      <div>
        <button onClick={() => localStorage.removeItem('token')} className="text-red-600">
          Logout
        </button>
      </div>
    </div>
  )
}
