import { useParams } from 'react-router'

export default function Account() {
  const { accountId } = useParams()
  return (
    <div className="p-8">
      <h1 className="text-2xl font-bold tracking-tight">Account #{accountId}</h1>
      <p className="mt-2 text-slate-400">The account summary and transactions will live here.</p>
    </div>
  )
}
