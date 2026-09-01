"use client"

import { useRouter } from "next/navigation"
import { type FormEvent, useState } from "react"
import { type Hash } from "viem"
import { waitForTransactionReceipt, writeContract } from "wagmi/actions"
import { useAccount, useChainId, useConfig, useSwitchChain } from "wagmi"

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import {
  ApiError,
  createMarket,
  localDemoUserId,
} from "@/lib/api"
import {
  ARC_TESTNET_CHAIN_ID,
  ARC_TESTNET_USDC_ADDRESS,
  SIGNAL_ARC_MARKET_FACTORY_ABI,
  SIGNAL_ARC_MARKET_FACTORY_ADDRESS,
  isArcTestnetChain,
} from "@/lib/contracts"
import { arcTestnet } from "@/lib/wagmi"
import {
  closeTimestampSeconds,
  CreateMarketAdvancedFields,
  CreateMarketCategoryField,
  CreateMarketCoverUploadField,
  CreateMarketFields,
  CreateMarketOutcomeFields,
  CreateMarketScheduleFields,
  CreateMarketSubmitStatus,
  getDeployedMarketAddress,
  getDeployErrorMessage,
  optionalText,
  requiredText,
  toRfc3339,
  type DeployState,
  type SubmitState,
} from "@/modules/markets/create"

function getErrorState(
  error: unknown,
): Extract<SubmitState, { status: "error" }> {
  if (error instanceof ApiError) {
    return {
      status: "error",
      message: error.message,
      requestId: error.requestId,
    }
  }

  if (error instanceof Error) {
    return {
      status: "error",
      message: error.message,
      requestId: null,
    }
  }

  return {
    status: "error",
    message: "Unable to create market.",
    requestId: null,
  }
}

export function CreateMarketForm() {
  const router = useRouter()
  const config = useConfig()
  const { address, isConnected } = useAccount()
  const chainId = useChainId()
  const { switchChain, isPending: isSwitchingChain } = useSwitchChain()
  const [state, setState] = useState<SubmitState>({ status: "idle" })
  const [deployState, setDeployState] = useState<DeployState>({
    status: "idle",
  })
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [coverImageUrl, setCoverImageUrl] = useState("")
  const [isUploadingImage, setIsUploadingImage] = useState(false)
  const isArcTestnet = isArcTestnetChain(chainId)
  const isDeploying = deployState.status === "deploying"
  const canSubmit = Boolean(
    isConnected &&
      isArcTestnet &&
      SIGNAL_ARC_MARKET_FACTORY_ADDRESS &&
      !isDeploying &&
      !isSwitchingChain,
  )

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    // Never submit while an image upload is mid-flight; the cover URL would
    // be stale or empty. The submit button is disabled in this state too.
    if (isUploadingImage) return

    if (!isConnected || !address) {
      setState({
        status: "error",
        message: "Connect a wallet before deploying and creating a market.",
        requestId: null,
      })
      return
    }

    if (!isArcTestnet) {
      setState({
        status: "error",
        message: "Switch to Arc Testnet before deploying and creating a market.",
        requestId: null,
      })
      return
    }

    if (!SIGNAL_ARC_MARKET_FACTORY_ADDRESS) {
      setState({
        status: "error",
        message: "Market factory address is not configured.",
        requestId: null,
      })
      return
    }

    setState({ status: "submitting" })
    setDeployState({ status: "idle" })

    const formData = new FormData(event.currentTarget)
    const opensAt = optionalText(formData, "opens_at")
    const closesAt = requiredText(formData, "closes_at")
    const marketId = crypto.randomUUID()
    const title = requiredText(formData, "title")
    const closesAtRfc3339 = toRfc3339(closesAt)

    let hash: Hash | undefined
    try {
      setDeployState({ status: "deploying", marketId, marketTitle: title })
      hash = await writeContract(config, {
        address: SIGNAL_ARC_MARKET_FACTORY_ADDRESS,
        abi: SIGNAL_ARC_MARKET_FACTORY_ABI,
        functionName: "createMarket",
        args: [
          marketId,
          title,
          closeTimestampSeconds(closesAtRfc3339),
          address,
          ARC_TESTNET_USDC_ADDRESS,
        ],
        chainId: ARC_TESTNET_CHAIN_ID,
        account: address,
      })
      setDeployState({ status: "deploying", hash, marketId, marketTitle: title })

      const receipt = await waitForTransactionReceipt(config, {
        hash,
        chainId: ARC_TESTNET_CHAIN_ID,
      })
      const marketAddress = getDeployedMarketAddress(receipt)

      const response = await createMarket({
        id: marketId,
        creator_user_id: requiredText(formData, "creator_user_id"),
        title,
        description: optionalText(formData, "description"),
        category: optionalText(formData, "category"),
        cover_image_url: coverImageUrl.trim() || undefined,
        outcome_yes_label: optionalText(formData, "outcome_yes_label"),
        outcome_no_label: optionalText(formData, "outcome_no_label"),
        collateral_asset: optionalText(formData, "collateral_asset"),
        chain: requiredText(formData, "chain"),
        resolution_source: optionalText(formData, "resolution_source"),
        opens_at: opensAt ? toRfc3339(opensAt) : undefined,
        closes_at: closesAtRfc3339,
        market_contract_address: marketAddress,
        market_deployment_tx_hash: hash,
        market_factory_address: SIGNAL_ARC_MARKET_FACTORY_ADDRESS,
        resolver_address: address,
      })

      setDeployState({
        status: "success",
        hash,
        marketAddress,
        marketId,
        marketTitle: title,
      })
      setState({ status: "success", market: response.data.market })
      router.push(`/markets/${marketId}`)
    } catch (error) {
      const message = getDeployErrorMessage(error)
      setDeployState({
        status: "error",
        message,
        hash,
        marketId,
        marketTitle: title,
      })
      if (error instanceof ApiError) {
        setState(getErrorState(error))
      } else {
        setState({
          status: "error",
          message,
          requestId: null,
        })
      }
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create a Market</CardTitle>
        <CardDescription>
          Deploy the Arc Testnet contract first; the market becomes public only after deployment succeeds.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form className="grid gap-5" onSubmit={handleSubmit}>
          <input type="hidden" name="creator_user_id" value={localDemoUserId} />

          <CreateMarketFields />

          <CreateMarketCoverUploadField
            coverImageUrl={coverImageUrl}
            onCoverImageUrlChange={setCoverImageUrl}
            onUploadingChange={setIsUploadingImage}
            disabled={state.status === "submitting"}
            isUploadingImage={isUploadingImage}
          />

          <CreateMarketCategoryField
            disabled={state.status === "submitting" || isUploadingImage}
          />

          <CreateMarketOutcomeFields />

          <CreateMarketScheduleFields />

          <CreateMarketAdvancedFields
            showAdvanced={showAdvanced}
            onToggleAdvanced={() => setShowAdvanced(!showAdvanced)}
          />

          <CreateMarketSubmitStatus
            state={state}
            deployState={deployState}
            isUploadingImage={isUploadingImage}
            canSubmit={canSubmit}
            isConnected={isConnected}
            isArcTestnet={isArcTestnet}
            isSwitchingChain={isSwitchingChain}
            onSwitchChain={() => switchChain({ chainId: arcTestnet.id })}
          />
        </form>
      </CardContent>
    </Card>
  )
}
