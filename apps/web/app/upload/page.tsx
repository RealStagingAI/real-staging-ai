"use client";

import { useEffect, useState, useCallback, useRef } from "react";
import { useRouter } from "next/navigation";
import { Upload as UploadIcon, FolderOpen, Plus, RefreshCw, CheckCircle2, Loader2, FileImage, X, AlertCircle, CreditCard, Lock } from "lucide-react";
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";

type Project = {
  id: string
  name: string
}

type ProjectListResponse = {
  projects: Project[]
}

type FileWithOverrides = {
  file: File
  id: string
  previewUrl: string
  roomType?: string
  styles?: string[]
  prompt?: string
}

type UploadProgress = {
  fileId: string
  status: 'pending' | 'presigning' | 'uploading' | 'creating' | 'success' | 'error'
  progress: number
  error?: string
  imageId?: string
}

type Subscription = {
  id: string
  status: string
}

type SubscriptionResponse = {
  items: Subscription[]
}

type UsageStats = {
  images_used: number
  monthly_limit: number
  plan_code: string
  period_start: string
  period_end: string
  has_subscription: boolean
  remaining_images: number
}

export default function UploadPage() {
  const router = useRouter()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [files, setFiles] = useState<FileWithOverrides[]>([])
  const [projectId, setProjectId] = useState("")
  const [defaultRoomType, setDefaultRoomType] = useState("")
  const [defaultStyles, setDefaultStyles] = useState<string[]>([])
  const [defaultPrompt, setDefaultPrompt] = useState("")
  const [status, setStatus] = useState<string>("")
  const [projects, setProjects] = useState<Project[]>([])
  const [newProjectName, setNewProjectName] = useState("")
  const [isUploading, setIsUploading] = useState(false)
  const [isDragging, setIsDragging] = useState(false)
  const [uploadProgress, setUploadProgress] = useState<Record<string, UploadProgress>>({})
  const [hasActiveSubscription, setHasActiveSubscription] = useState<boolean | null>(null)
  const [subscriptionLoading, setSubscriptionLoading] = useState(true)
  const [usage, setUsage] = useState<UsageStats | null>(null)
  const [usageLoading, setUsageLoading] = useState(true)

  async function checkSubscriptionAndUsage() {
    try {
      setSubscriptionLoading(true)
      setUsageLoading(true)
      
      const [subsRes, usageRes] = await Promise.all([
        apiFetch<SubscriptionResponse>("/v1/billing/subscriptions"),
        apiFetch<UsageStats>("/v1/billing/usage")
      ])
      
      const activeSubscription = subsRes.items?.some(
        sub => sub.status === "active" || sub.status === "trialing"
      )
      setHasActiveSubscription(activeSubscription || false)
      setUsage(usageRes)
    } catch (err: unknown) {
      console.error("Failed to check subscription and usage:", err)
      setHasActiveSubscription(false)
    } finally {
      setSubscriptionLoading(false)
      setUsageLoading(false)
    }
  }

  async function loadProjects() {
    try {
      setStatus("Loading projects...")
      const res = await apiFetch<ProjectListResponse>("/v1/projects")
      setProjects(res.projects || [])
      if (!projectId && res.projects && res.projects.length > 0) {
        setProjectId(res.projects[0].id)
      }
      setStatus("")
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err)
      setStatus(message)
    }
  }

  async function createProject() {
    if (!newProjectName.trim()) {
      setStatus("Please provide a project name.")
      return
    }
    try {
      setStatus("Creating project...")
      const created = await apiFetch<Project>("/v1/projects", {
        method: "POST",
        body: JSON.stringify({ name: newProjectName.trim() }),
      })
      setNewProjectName("")
      await loadProjects()
      setProjectId(created.id)
      setStatus("Project created.")
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err)
      setStatus(message)
    }
  }

  useEffect(() => {
    checkSubscriptionAndUsage()
    loadProjects()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // Cleanup preview URLs on unmount
  useEffect(() => {
    return () => {
      files.forEach(file => {
        URL.revokeObjectURL(file.previewUrl)
      })
    }
  }, [files])

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
    const droppedFiles = Array.from(e.dataTransfer.files).filter(f => f.type.startsWith('image/'))
    if (droppedFiles.length > 0) {
      addFiles(droppedFiles)
    }
  }, [])

  const handleFileSelect = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = e.target.files ? Array.from(e.target.files) : []
    if (selectedFiles.length > 0) {
      addFiles(selectedFiles)
    }
  }, [])

  const addFiles = (newFiles: File[]) => {
    const filesWithData: FileWithOverrides[] = newFiles.map(file => ({
      file,
      id: `${Date.now()}-${Math.random()}`,
      previewUrl: URL.createObjectURL(file),
    }))
    setFiles(prev => [...prev, ...filesWithData])
  }

  const removeFile = (fileId: string) => {
    setFiles(prev => {
      const fileToRemove = prev.find(f => f.id === fileId)
      // Clean up the preview URL to avoid memory leaks
      if (fileToRemove) {
        URL.revokeObjectURL(fileToRemove.previewUrl)
      }
      return prev.filter(f => f.id !== fileId)
    })
    setUploadProgress(prev => {
      const newProgress = { ...prev }
      delete newProgress[fileId]
      return newProgress
    })
  }

  const updateFileOverride = (fileId: string, field: 'roomType' | 'styles', value: string | string[]) => {
    setFiles(prev => prev.map(f => {
      if (f.id === fileId) {
        if (field === 'styles') {
          return { ...f, styles: Array.isArray(value) ? value : (value ? [value] : undefined) }
        }
        // roomType is always string
        return { ...f, roomType: typeof value === 'string' ? (value || undefined) : undefined }
      }
      return f
    }))
  }

  async function uploadSingleFile(fileData: FileWithOverrides): Promise<{ success: boolean; imageId?: string; error?: string }> {
    const updateProgress = (status: UploadProgress['status'], progress: number, error?: string) => {
      setUploadProgress(prev => ({
        ...prev,
        [fileData.id]: { fileId: fileData.id, status, progress, error }
      }))
    }

    try {
      // 1) Presign
      updateProgress('presigning', 10)
      const presign = await apiFetch<{ upload_url: string; file_key: string }>(
        "/v1/uploads/presign",
        {
          method: "POST",
          body: JSON.stringify({
            filename: fileData.file.name,
            content_type: fileData.file.type || "application/octet-stream",
            file_size: fileData.file.size,
          }),
        }
      ).catch((err) => {
        // Handle subscription_required error specifically
        if (err.message && err.message.includes('subscription_required')) {
          throw new Error('An active subscription is required to upload images. Please subscribe to continue.');
        }
        throw err;
      })

      // 2) Upload to S3
      updateProgress('uploading', 40)
      const putRes = await fetch(presign.upload_url, {
        method: "PUT",
        headers: {
          "Content-Type": fileData.file.type || "application/octet-stream",
          "Cache-Control": "public, max-age=31536000, immutable",
        },
        body: fileData.file,
      })
      if (!putRes.ok) {
        throw new Error(`Upload failed: ${putRes.status}`)
      }

      // 3) Create Image
      updateProgress('creating', 70)
      const u = new URL(presign.upload_url)
      const originalUrl = `${u.origin}${u.pathname}`

      const roomType = fileData.roomType || defaultRoomType
      const styles = fileData.styles && fileData.styles.length > 0 ? fileData.styles : defaultStyles
      const prompt = fileData.prompt || defaultPrompt

      // If multiple styles selected, use batch endpoint
      if (styles.length > 1) {
        const images = styles.map(style => {
          const body: { project_id: string; original_url: string; room_type?: string; style?: string; prompt?: string } = {
            project_id: projectId,
            original_url: originalUrl,
          }
          if (roomType) body.room_type = roomType
          if (style) body.style = style
          if (prompt && prompt.length >= 10) body.prompt = prompt
          return body
        })

        const batchResponse = await apiFetch<{ images: Array<{ id: string }>, success: number, failed: number }>("/v1/images/batch", {
          method: "POST",
          body: JSON.stringify({ images }),
        })
        
        if (batchResponse.failed > 0) {
          throw new Error(`${batchResponse.failed} of ${styles.length} variants failed to create`)
        }
        
        updateProgress('success', 100)
        const firstImageId = batchResponse.images[0]?.id
        if (firstImageId) {
          setUploadProgress(prev => ({
            ...prev,
            [fileData.id]: { ...prev[fileData.id], imageId: firstImageId }
          }))
        }
        return { success: true, imageId: firstImageId }
      } else {
        // Single style or no style - use single endpoint
        const body: { project_id: string; original_url: string; room_type?: string; style?: string; prompt?: string } = {
          project_id: projectId,
          original_url: originalUrl,
        }
        if (roomType) body.room_type = roomType
        if (styles.length > 0) body.style = styles[0]
        if (prompt && prompt.length >= 10) body.prompt = prompt

        const created = await apiFetch<{ id: string }>("/v1/images", {
          method: "POST",
          body: JSON.stringify(body),
        })
        
        updateProgress('success', 100)
        setUploadProgress(prev => ({
          ...prev,
          [fileData.id]: { ...prev[fileData.id], imageId: created.id }
        }))
        
        return { success: true, imageId: created.id }
      }

    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err)
      updateProgress('error', 0, message)
      return { success: false, error: message }
    }
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setStatus("")
    setIsUploading(true)
    
    if (files.length === 0) {
      setStatus("Please select at least one file.")
      setIsUploading(false)
      return
    }
    if (!projectId) {
      setStatus("Please select or create a project.")
      setIsUploading(false)
      return
    }
    
    // Check usage limit
    if (usage && usage.remaining_images <= 0) {
      setStatus("You have reached your monthly image limit. Please upgrade your plan to continue.")
      setIsUploading(false)
      return
    }
    
    // Warn if uploading more than remaining
    if (usage && files.length > usage.remaining_images) {
      setStatus(`Warning: You are trying to upload ${files.length} images but only have ${usage.remaining_images} remaining this month. Only ${usage.remaining_images} will be processed.`)
    }

    // Upload all files concurrently
    const results = await Promise.all(files.map(uploadSingleFile))
    
    const successCount = results.filter(r => r.success).length
    const errorCount = results.filter(r => !r.success).length
    
    if (errorCount === 0) {
      setStatus(`Success! ${successCount} image${successCount > 1 ? 's' : ''} uploaded and queued for staging.`)
      // Reset after delay
      setTimeout(() => {
        setFiles([])
        setUploadProgress({})
        setDefaultRoomType("")
        setDefaultStyles([])
      }, 3000)
    } else if (successCount === 0) {
      setStatus(`Upload failed for all ${errorCount} images. See individual errors above.`)
    } else {
      setStatus(`Partial success: ${successCount} succeeded, ${errorCount} failed. See details above.`)
    }
    
    setIsUploading(false)
  }

  const successfulUploads = Object.values(uploadProgress).filter(p => p.status === 'success')

  const isAtLimit = !usageLoading && usage && usage.remaining_images <= 0
  const canUpload = !subscriptionLoading && !usageLoading && hasActiveSubscription !== false && !isAtLimit

  return (
    <div className="container max-w-7xl py-4 sm:py-6 lg:py-8 space-y-4 sm:space-y-6 lg:space-y-8">
      {/* Usage Limit Reached Banner */}
      {isAtLimit && (
        <div className="bg-gradient-to-r from-red-50 to-orange-50 dark:from-red-950/20 dark:to-orange-950/20 border-2 border-red-200 dark:border-red-800 rounded-xl p-4 sm:p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row items-start gap-3 sm:gap-4">
            <div className="flex-shrink-0">
              <AlertCircle className="h-6 w-6 sm:h-8 sm:w-8 text-red-600 dark:text-red-400" />
            </div>
            <div className="flex-1 space-y-2">
              <h3 className="text-base sm:text-lg font-semibold text-red-900 dark:text-red-300">
                Monthly Limit Reached
              </h3>
              <p className="text-sm sm:text-base text-red-800 dark:text-red-400">
                You have used all {usage?.monthly_limit} images in your {usage?.plan_code?.toUpperCase() || 'FREE'} plan for this billing period. Upgrade to continue staging more properties.
              </p>
              <div className="flex flex-col sm:flex-row gap-2 sm:gap-3 mt-3">
                <button
                  onClick={() => router.push('/profile')}
                  className="inline-flex items-center justify-center gap-2 px-4 py-2.5 sm:py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-lg transition-colors touch-manipulation w-full sm:w-auto"
                >
                  <CreditCard className="h-4 w-4" />
                  Upgrade Plan
                </button>
                <button
                  onClick={() => router.push('/billing')}
                  className="inline-flex items-center justify-center gap-2 px-4 py-2.5 sm:py-2 bg-white dark:bg-gray-800 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 font-medium rounded-lg hover:bg-red-50 dark:hover:bg-red-950/30 transition-colors touch-manipulation w-full sm:w-auto"
                >
                  View Usage
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Subscription Required Banner */}
      {!subscriptionLoading && !usageLoading && hasActiveSubscription === false && !isAtLimit && (
        <div className="bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-950/20 dark:to-orange-950/20 border-2 border-amber-200 dark:border-amber-800 rounded-xl p-4 sm:p-6 shadow-sm">
          <div className="flex flex-col sm:flex-row items-start gap-3 sm:gap-4">
            <div className="flex-shrink-0">
              <Lock className="h-6 w-6 sm:h-8 sm:w-8 text-amber-600 dark:text-amber-400" />
            </div>
            <div className="flex-1 space-y-2">
              <h3 className="text-base sm:text-lg font-semibold text-amber-900 dark:text-amber-300">
                Subscription Required
              </h3>
              <p className="text-sm sm:text-base text-amber-800 dark:text-amber-400">
                An active subscription is required to upload and process images. Upgrade your plan to unlock 500 images and priority support.
              </p>
              <button
                onClick={() => router.push('/profile')}
                className="mt-3 inline-flex items-center justify-center gap-2 px-4 py-2.5 sm:py-2 bg-amber-600 hover:bg-amber-700 text-white font-medium rounded-lg transition-colors touch-manipulation w-full sm:w-auto"
              >
                <CreditCard className="h-4 w-4" />
                View Subscription Plans
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Header */}
      <div>
        <h1 className="text-2xl sm:text-3xl font-bold mb-2">
          <span className="gradient-text">Upload & Stage</span>
        </h1>
        <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400">
          Upload multiple property photos and transform them with AI-powered virtual staging
        </p>
      </div>

      {/* Project Management */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <FolderOpen className="h-5 w-5 text-blue-600" />
            <span>Project Selection</span>
          </div>
        </div>
        <div className="card-body space-y-4">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Create New Project</label>
              <input
                className="input"
                value={newProjectName}
                onChange={(e) => setNewProjectName(e.target.value)}
                placeholder="e.g., Downtown Condo Staging"
                onKeyDown={(e) => e.key === 'Enter' && createProject()}
              />
            </div>
            <button 
              type="button" 
              className="btn btn-secondary sm:mt-7 w-full sm:w-auto" 
              onClick={createProject}
              disabled={!newProjectName.trim()}
            >
              <Plus className="h-4 w-4" />
              Create
            </button>
          </div>

          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Select Project
              </label>
              <select
                className="input"
                value={projectId}
                onChange={(e) => setProjectId(e.target.value)}
              >
                <option value="">Choose a project...</option>
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
            <button 
              type="button" 
              className="btn btn-ghost sm:mt-7 w-full sm:w-auto" 
              onClick={loadProjects}
            >
              <RefreshCw className="h-4 w-4" />
              Refresh
            </button>
          </div>
        </div>
      </div>

      {/* Upload Form */}
      <form onSubmit={onSubmit} className="card">
        <div className="card-header">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <UploadIcon className="h-5 w-5 text-blue-600" />
              <span>Upload Images</span>
            </div>
            {files.length > 0 && (
              <span className="text-sm text-gray-600 dark:text-gray-400">
                {files.length} file{files.length > 1 ? 's' : ''} selected
              </span>
            )}
          </div>
        </div>
        <div className="card-body space-y-6">
          {/* Drag and Drop Zone */}
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">Property Images</label>
            <div
              onClick={() => canUpload && fileInputRef.current?.click()}
              onDragOver={canUpload ? handleDragOver : undefined}
              onDragLeave={canUpload ? handleDragLeave : undefined}
              onDrop={canUpload ? handleDrop : undefined}
              className={cn(
                "relative rounded-xl border-2 border-dashed transition-all duration-200 p-6 sm:p-8",
                !canUpload ? "opacity-50 cursor-not-allowed border-gray-200 dark:border-gray-800" :
                isDragging 
                  ? "border-blue-500 bg-blue-50 dark:border-blue-400 dark:bg-blue-950/30 cursor-pointer" 
                  : "border-gray-300 hover:border-gray-400 dark:border-gray-600 dark:hover:border-gray-500 cursor-pointer",
                files.length > 0 && "border-green-500 bg-green-50/30 dark:border-green-500 dark:bg-green-950/30"
              )}
            >
              <input
                ref={fileInputRef}
                type="file"
                multiple
                accept="image/*"
                className="hidden"
                onChange={handleFileSelect}
                disabled={!canUpload}
              />
              <div className="flex flex-col items-center justify-center text-center space-y-3">
                <div className={cn(
                  "rounded-xl p-3",
                  files.length > 0 ? "bg-green-100 dark:bg-green-900/50" : "bg-blue-100 dark:bg-blue-900/50"
                )}>
                  <FileImage className={cn(
                    "h-6 w-6 sm:h-8 sm:w-8",
                    files.length > 0 ? "text-green-600 dark:text-green-500" : "text-blue-600 dark:text-blue-500"
                  )} />
                </div>
                <div>
                  <p className="font-medium text-gray-900 dark:text-gray-100">
                    {files.length > 0 
                      ? `${files.length} file${files.length > 1 ? 's' : ''} ready to upload`
                      : "Drag & drop your images here"
                    }
                  </p>
                  <p className="text-sm text-gray-600 dark:text-gray-400 mt-1">
                    or click to browse • Max 10MB per file
                  </p>
                </div>
                <p className="text-xs text-gray-500 dark:text-gray-400">
                  Supports: JPG, PNG, WEBP • Upload multiple files at once
                </p>
              </div>
            </div>
          </div>

          {/* Default Staging Options - Show at top when files selected */}
          {files.length > 0 && (
            <div className="space-y-3">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">
                Default Settings <span className="text-gray-400 dark:text-gray-500 font-normal">(applied to all images using &quot;Use Default&quot;)</span>
              </h3>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Room Type <span className="text-gray-400 dark:text-gray-500 font-normal">(optional)</span>
                  </label>
                  <select
                    className="input"
                    value={defaultRoomType}
                    onChange={(e) => setDefaultRoomType(e.target.value)}
                    disabled={isUploading}
                  >
                    <option value="">Auto-detect</option>
                    <option value="living_room">Living Room</option>
                    <option value="bedroom">Bedroom</option>
                    <option value="kitchen">Kitchen</option>
                    <option value="bathroom">Bathroom</option>
                    <option value="dining_room">Dining Room</option>
                    <option value="office">Office</option>
                    <option value="entryway">Entryway</option>
                    <option value="outdoor">Outdoor/Patio</option>
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Furniture Styles <span className="text-gray-400 dark:text-gray-500 font-normal">(select multiple)</span>
                  </label>
                  <div className="space-y-2 p-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-800">
                    {[
                      { value: 'modern', label: 'Modern' },
                      { value: 'contemporary', label: 'Contemporary' },
                      { value: 'traditional', label: 'Traditional' },
                      { value: 'industrial', label: 'Industrial' },
                      { value: 'scandinavian', label: 'Scandinavian' }
                    ].map(({ value, label }) => (
                      <label key={value} className="flex items-center space-x-2 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-700 p-1 rounded">
                        <input
                          type="checkbox"
                          checked={defaultStyles.includes(value)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setDefaultStyles([...defaultStyles, value])
                            } else {
                              setDefaultStyles(defaultStyles.filter(s => s !== value))
                            }
                          }}
                          disabled={isUploading}
                          className="rounded border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                        />
                        <span className="text-sm text-gray-700 dark:text-gray-300">{label}</span>
                      </label>
                    ))}
                  </div>
                  {defaultStyles.length > 0 && (
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                      {defaultStyles.length} style{defaultStyles.length > 1 ? 's' : ''} selected · Creates {defaultStyles.length} variant{defaultStyles.length > 1 ? 's' : ''} per image
                    </p>
                  )}
                </div>
              </div>
              
              {/* Custom Prompt (Testing) */}
              <div className="mt-4">
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Custom Prompt <span className="text-xs text-gray-500">(optional - for testing)</span>
                </label>
                <textarea
                  className="input min-h-[100px] font-mono text-xs"
                  value={defaultPrompt}
                  onChange={(e) => setDefaultPrompt(e.target.value)}
                  placeholder="Leave empty to use library prompts. Enter custom prompt for testing (10-2000 characters)..."
                  disabled={isUploading}
                  rows={4}
                />
                <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">
                  Custom prompts override the built-in prompt library for the selected room/style combination.
                </p>
              </div>
            </div>
          )}

          {/* File List with Individual Settings */}
          {files.length > 0 && (
            <div className="space-y-3">
              <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300">Selected Files</h3>
              <div className="space-y-3 max-h-[60vh] sm:max-h-96 overflow-y-auto scroll-smooth-mobile">
                {files.map((fileData) => {
                  const progress = uploadProgress[fileData.id]
                  return (
                    <div 
                      key={fileData.id} 
                      className={cn(
                        "border rounded-lg p-3 sm:p-4 transition-all",
                        progress?.status === 'success' && "border-green-300 bg-green-50 dark:border-green-800 dark:bg-green-950/30",
                        progress?.status === 'error' && "border-red-300 bg-red-50 dark:border-red-800 dark:bg-red-950/30",
                        !progress && "border-gray-200 bg-white dark:border-gray-700 dark:bg-gray-800/50"
                      )}
                    >
                      <div className="flex items-start gap-3 sm:gap-4">
                        {/* Image Preview with Status Overlay */}
                        <div className="flex-shrink-0 relative">
                          {/* eslint-disable-next-line @next/next/no-img-element */}
                          <img 
                            src={fileData.previewUrl} 
                            alt={fileData.file.name}
                            className="w-16 h-16 sm:w-20 sm:h-20 object-cover rounded-lg border-2 border-gray-200 dark:border-gray-600"
                          />
                          {progress && (
                            <div className="absolute inset-0 flex items-center justify-center bg-black/50 rounded-lg">
                              {progress.status === 'success' && (
                                <CheckCircle2 className="h-6 w-6 sm:h-8 sm:w-8 text-green-400" />
                              )}
                              {progress.status === 'error' && (
                                <AlertCircle className="h-6 w-6 sm:h-8 sm:w-8 text-red-400" />
                              )}
                              {!['success', 'error'].includes(progress.status) && (
                                <Loader2 className="h-6 w-6 sm:h-8 sm:w-8 text-blue-400 animate-spin" />
                              )}
                            </div>
                          )}
                        </div>

                        <div className="flex-1 min-w-0">
                          <div className="flex items-start justify-between gap-2">
                            <div className="flex-1 min-w-0">
                              <p className="font-medium text-gray-900 dark:text-gray-100 truncate">{fileData.file.name}</p>
                              <p className="text-xs text-gray-600 dark:text-gray-400 mt-0.5">
                                {(fileData.file.size / 1024 / 1024).toFixed(2)} MB
                              </p>
                            </div>
                            {!isUploading && (
                              <button
                                type="button"
                                onClick={() => removeFile(fileData.id)}
                                className="text-gray-400 hover:text-red-600 dark:text-gray-500 dark:hover:text-red-500 transition-colors"
                              >
                                <X className="h-4 w-4" />
                              </button>
                            )}
                          </div>

                          {/* Progress Bar */}
                          {progress && progress.status !== 'success' && progress.status !== 'error' && (
                            <div className="mt-2">
                              <div className="flex items-center justify-between text-xs text-gray-600 dark:text-gray-400 mb-1">
                                <span className="capitalize">{progress.status}...</span>
                                <span>{progress.progress}%</span>
                              </div>
                              <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-1.5">
                                <div 
                                  className="bg-blue-600 dark:bg-blue-500 h-1.5 rounded-full transition-all duration-300"
                                  style={{ width: `${progress.progress}%` }}
                                />
                              </div>
                            </div>
                          )}

                          {/* Error Message */}
                          {progress?.status === 'error' && progress.error && (
                            <p className="text-sm text-red-600 dark:text-red-400 mt-2">{progress.error}</p>
                          )}

                          {/* Success Message */}
                          {progress?.status === 'success' && progress.imageId && (
                            <p className="text-sm text-green-600 dark:text-green-400 mt-2">
                              Successfully uploaded! Image ID: {progress.imageId}
                            </p>
                          )}

                          {/* Always show settings inline (not collapsed) */}
                          {!isUploading && (
                            <div className="mt-3 grid grid-cols-1 sm:grid-cols-2 gap-3">
                              <div>
                                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                                  Room Type
                                </label>
                                <select
                                  className="input text-sm"
                                  value={fileData.roomType || ''}
                                  onChange={(e) => updateFileOverride(fileData.id, 'roomType', e.target.value)}
                                >
                                  <option value="">Use Default</option>
                                  <option value="living_room">Living Room</option>
                                  <option value="bedroom">Bedroom</option>
                                  <option value="kitchen">Kitchen</option>
                                  <option value="bathroom">Bathroom</option>
                                  <option value="dining_room">Dining Room</option>
                                  <option value="office">Office</option>
                                  <option value="entryway">Entryway</option>
                                  <option value="outdoor">Outdoor/Patio</option>
                                </select>
                              </div>
                              <div>
                                <label className="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
                                  Style Override
                                </label>
                                <select
                                  className="input text-sm"
                                  value={fileData.styles?.[0] || ''}
                                  onChange={(e) => updateFileOverride(fileData.id, 'styles', e.target.value)}
                                >
                                  <option value="">Use Default</option>
                                  <option value="modern">Modern (single)</option>
                                  <option value="contemporary">Contemporary (single)</option>
                                  <option value="traditional">Traditional (single)</option>
                                  <option value="industrial">Industrial (single)</option>
                                  <option value="scandinavian">Scandinavian (single)</option>
                                </select>
                              </div>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}

          {/* Submit Button */}
          <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 sm:gap-4 pt-3 sm:pt-4 border-t">
            <button 
              className="btn btn-primary w-full sm:w-auto order-1 sm:order-none" 
              type="submit"
              disabled={isUploading || files.length === 0 || !projectId || !canUpload}
            >
              {isUploading ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Uploading {files.length} image{files.length > 1 ? 's' : ''}...
                </>
              ) : (
                <>
                  <UploadIcon className="h-4 w-4" />
                  Upload & Stage {files.length > 0 ? `${files.length} Image${files.length > 1 ? 's' : ''}` : 'Images'}
                </>
              )}
            </button>
            
            {status && (
              <div className={cn(
                "text-xs sm:text-sm font-medium px-3 sm:px-4 py-2 rounded-lg order-2 sm:order-none",
                status.includes("Success") || status.includes("succeeded")
                  ? "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400"
                  : status.includes("failed") || status.includes("error")
                  ? "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400"
                  : "bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400"
              )}>
                {status}
              </div>
            )}
          </div>
        </div>
      </form>

      {/* Success Summary */}
      {successfulUploads.length > 0 && (
        <div className="card border-green-200 bg-green-50/50 dark:border-green-800 dark:bg-green-950/30 animate-in">
          <div className="card-body">
            <div className="flex items-start gap-4">
              <div className="rounded-xl bg-green-100 dark:bg-green-900/50 p-2">
                <CheckCircle2 className="h-6 w-6 text-green-600 dark:text-green-500" />
              </div>
              <div className="flex-1">
                <h3 className="font-semibold text-green-900 dark:text-green-100 mb-1">
                  {successfulUploads.length} Image{successfulUploads.length > 1 ? 's' : ''} Successfully Queued!
                </h3>
                <p className="text-sm text-green-700 dark:text-green-300 mb-3">
                  Your images have been uploaded and are being processed by our AI staging system.
                </p>
                <a 
                  href="/images" 
                  className="inline-flex items-center gap-2 text-sm font-medium text-green-700 dark:text-green-400 hover:text-green-800 dark:hover:text-green-300"
                >
                  View in Images Dashboard →
                </a>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
