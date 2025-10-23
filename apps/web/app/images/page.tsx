"use client";

import { useEffect, useMemo, useState, useCallback, useRef } from "react";
import JSZip from "jszip";
import NextImage from "next/image";

import { 
  FolderOpen, 
  RefreshCw, 
  Grid3x3, 
  List, 
  Download, 
  CheckCircle2,
  ExternalLink,
  Loader2,
  Image as ImageIcon,
  AlertCircle,
  Check,
  X,
  Trash2,
  ChevronLeft,
  ChevronRight,
  XCircle
} from "lucide-react";

import { apiFetch } from "@/lib/api";
import { cn, formatRelativeTime } from "@/lib/utils";
import { getCachedUrl, setCachedUrl, clearExpiredCache } from "@/lib/imageCache";

type Project = {
  id: string;
  name: string;
};

type ProjectListResponse = {
  projects: Project[];
};

type ImageRecord = {
  id: string;
  project_id: string;
  original_url: string;
  staged_url?: string | null;
  status: string;
  error?: string | null;
  room_type?: string | null;
  style?: string | null;
  seed?: number | null;
  created_at: string;
  updated_at: string;
};

type ImageListResponse = {
  images: ImageRecord[];
};

export default function ImagesPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProjectId, setSelectedProjectId] = useState<string>("");
  const [images, setImages] = useState<ImageRecord[]>([]);
  const [selectedImageIds, setSelectedImageIds] = useState<Set<string>>(new Set());
  const [focusedImageId, setFocusedImageId] = useState<string | null>(null);
  const [statusMessage, setStatusMessage] = useState<string>("");
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [loadingImages, setLoadingImages] = useState(false);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [imageUrls, setImageUrls] = useState<Record<string, { original?: string; staged?: string }>>({});
  const [hoveredImageId, setHoveredImageId] = useState<string | null>(null);
  const [downloadType, setDownloadType] = useState<'original' | 'staged'>('staged');
  const [pollingInterval, setPollingInterval] = useState<NodeJS.Timeout | null>(null);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const imagesRef = useRef<ImageRecord[]>([]);
  const pollingStartTimeRef = useRef<number | null>(null);
  
  // Preview/Lightbox state
  const [previewImageId, setPreviewImageId] = useState<string | null>(null);
  const [previewMode, setPreviewMode] = useState<'original' | 'staged'>('staged');

  // Lazy loading state
  const [visibleImageIds, setVisibleImageIds] = useState<Set<string>>(new Set());
  const visibleImageIdsRef = useRef<Set<string>>(new Set());
  const imageObserverRef = useRef<IntersectionObserver | null>(null);

  // Keep refs in sync with state
  useEffect(() => {
    imagesRef.current = images;
  }, [images]);
  
  useEffect(() => {
    visibleImageIdsRef.current = visibleImageIds;
  }, [visibleImageIds]);
  
  // Clear expired cache entries on mount
  useEffect(() => {
    clearExpiredCache();
  }, []);
  
  // Set up Intersection Observer for lazy loading
  useEffect(() => {
    if (typeof window === 'undefined') return;
    
    imageObserverRef.current = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            const imageId = entry.target.getAttribute('data-image-id');
            if (imageId) {
              setVisibleImageIds(prev => {
                // Only create new Set if imageId not already present
                if (prev.has(imageId)) {
                  return prev;
                }
                return new Set(prev).add(imageId);
              });
            }
          }
        });
      },
      {
        rootMargin: '50px', // Start loading 50px before entering viewport
        threshold: 0.01
      }
    );
    
    return () => {
      imageObserverRef.current?.disconnect();
    };
  }, []);

  // Register image elements with Intersection Observer
  const registerImageObserver = useCallback((element: HTMLElement | null, imageId: string) => {
    if (!element || !imageObserverRef.current) return;
    
    element.setAttribute('data-image-id', imageId);
    imageObserverRef.current.observe(element);
    
    return () => {
      imageObserverRef.current?.unobserve(element);
    };
  }, []);

  const selectedProject = useMemo(
    () => projects.find((project) => project.id === selectedProjectId) ?? null,
    [projects, selectedProjectId]
  );

  const toggleImageSelection = useCallback((imageId: string) => {
    setSelectedImageIds(prev => {
      const newSet = new Set(prev);
      if (newSet.has(imageId)) {
        newSet.delete(imageId);
      } else {
        newSet.add(imageId);
      }
      return newSet;
    });
  }, []);

  const selectAll = useCallback(() => {
    setSelectedImageIds(new Set(images.map(img => img.id)));
  }, [images]);

  const clearSelection = useCallback(() => {
    setSelectedImageIds(new Set());
  }, []);

  // Preview navigation
  const previewImages = useMemo(() => {
    // If images are selected, only preview those
    if (selectedImageIds.size > 0) {
      return images.filter(img => selectedImageIds.has(img.id));
    }
    // Otherwise preview all images
    return images;
  }, [images, selectedImageIds]);

  const currentPreviewIndex = useMemo(() => {
    if (!previewImageId) return -1;
    return previewImages.findIndex(img => img.id === previewImageId);
  }, [previewImageId, previewImages]);

  const goToPreviousImage = useCallback(() => {
    if (currentPreviewIndex > 0) {
      setPreviewImageId(previewImages[currentPreviewIndex - 1].id);
    }
  }, [currentPreviewIndex, previewImages]);

  const goToNextImage = useCallback(() => {
    if (currentPreviewIndex < previewImages.length - 1) {
      setPreviewImageId(previewImages[currentPreviewIndex + 1].id);
    }
  }, [currentPreviewIndex, previewImages]);

  const togglePreviewMode = useCallback(() => {
    setPreviewMode(prev => prev === 'original' ? 'staged' : 'original');
  }, []);

  const closePreview = useCallback(() => {
    setPreviewImageId(null);
  }, []);

  // Keyboard navigation for preview
  useEffect(() => {
    if (!previewImageId) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'ArrowLeft') {
        e.preventDefault();
        goToPreviousImage();
      } else if (e.key === 'ArrowRight') {
        e.preventDefault();
        goToNextImage();
      } else if (e.key === 'Escape') {
        e.preventDefault();
        closePreview();
      } else if (e.key === ' ' || e.key === 't' || e.key === 'T') {
        e.preventDefault();
        togglePreviewMode();
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [previewImageId, goToPreviousImage, goToNextImage, closePreview, togglePreviewMode]);

  // Get presigned S3 URL for viewing (with caching)
  async function getPresignedUrl(imageId: string, kind: 'original' | 'staged'): Promise<string | null> {
    // Check cache first
    const cached = getCachedUrl(imageId, kind);
    if (cached) {
      return cached;
    }
    
    try {
      // Fetch presigned URL from API with authentication
      const params = new URLSearchParams({ kind });
      const data = await apiFetch<{ url: string }>(`/v1/images/${imageId}/presign?${params.toString()}`);
      
      if (!data?.url) {
        console.error('No URL in presign response:', data);
        return null;
      }
      
      // Cache the presigned URL (valid for 1 hour typically)
      setCachedUrl(imageId, kind, data.url);
      
      return data.url;
    } catch (err: unknown) {
      console.error('Failed to get presigned URL:', err);
      return null;
    }
  }

  // Open image in preview lightbox
  function openPreview(imageId: string, kind: 'original' | 'staged') {
    setPreviewImageId(imageId);
    setPreviewMode(kind);
  }

  // Prefetch image URLs for display
  const prefetchImageUrls = useCallback(async (imageList: ImageRecord[]) => {
    // Use functional update to get current URLs without depending on imageUrls
    let currentUrls: Record<string, { original?: string; staged?: string }> = {};
    setImageUrls(prev => {
      currentUrls = prev;
      return prev;
    });
    
    // Get current visible IDs from ref to avoid stale closures
    const currentVisibleIds = visibleImageIdsRef.current;
    
    // Fetch URLs for:
    // 1. First 6 images (eager load for initial page)
    // 2. Any visible images (lazy load as user scrolls)
    // 3. Only if uploaded and not already cached
    const imagesToFetch = imageList.filter((img, index) => {
      const isFirstPage = index < 6; // Eagerly load first page
      const isVisible = currentVisibleIds.has(img.id);
      const isReady = img.status !== 'queued' && img.status !== 'processing';
      const needsFetch = !currentUrls[img.id]?.original || (img.staged_url && !currentUrls[img.id]?.staged);
      
      return (isFirstPage || isVisible) && isReady && needsFetch;
    });
    
    // Skip if no new images to fetch
    if (imagesToFetch.length === 0) {
      return;
    }
    
    // Fetch all URLs in parallel (with throttling to avoid overwhelming the API)
    const chunks = [];
    for (let i = 0; i < imagesToFetch.length; i += 5) {
      chunks.push(imagesToFetch.slice(i, i + 5));
    }

    const urlMap: Record<string, { original?: string; staged?: string }> = { ...currentUrls };
    
    for (const chunk of chunks) {
      await Promise.all(
        chunk.map(async (image) => {
          // Only fetch if we don't have it cached
          const originalUrl = urlMap[image.id]?.original || await getPresignedUrl(image.id, 'original');
          const stagedUrl = image.staged_url 
            ? (urlMap[image.id]?.staged || await getPresignedUrl(image.id, 'staged'))
            : null;
          
          urlMap[image.id] = {
            original: originalUrl || undefined,
            staged: stagedUrl || undefined
          };
        })
      );
    }

    setImageUrls(urlMap);
  }, []);
  
  // Trigger prefetch when images become visible
  useEffect(() => {
    if (visibleImageIds.size > 0 && images.length > 0) {
      prefetchImageUrls(images);
    }
  }, [visibleImageIds, images, prefetchImageUrls]);

  // Prefetch image on hover (for smooth transitions)
  const handleImageHover = useCallback((imageId: string) => {
    setHoveredImageId(imageId);
    
    const urls = imageUrls[imageId];
    if (urls?.original && typeof window !== "undefined") {
      const existing = document.querySelector(`img[src="${urls.original}"]`);
      if (!existing) {
        const preloadImg = new window.Image();
        preloadImg.src = urls.original;
      }
    }
  }, [imageUrls]);

  // Download image to local file system
  async function downloadImage(imageId: string, kind: 'original' | 'staged') {
    try {
      // Get presigned URL with download parameter
      const params = new URLSearchParams({ kind, download: '1' });
      const res = await apiFetch<{ url: string }>(`/v1/images/${imageId}/presign?${params.toString()}`);
      
      if (res?.url) {
        // Create a temporary anchor element to trigger download
        const link = document.createElement('a');
        link.href = res.url;
        link.download = ''; // Let browser determine filename from Content-Disposition header
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
      }
    } catch (err: unknown) {
      console.error('Failed to download image:', err);
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(`Download failed: ${message}`);
    }
  }

  async function downloadSelected() {
    const count = selectedImageIds.size;
    
    // Single file - download directly
    if (count === 1) {
      const imageId = Array.from(selectedImageIds)[0];
      const image = images.find(img => img.id === imageId);
      
      if (downloadType === 'staged' && image?.staged_url) {
        await downloadImage(imageId, 'staged');
      } else if (downloadType === 'original') {
        await downloadImage(imageId, 'original');
      }
      return;
    }
    
    // Multiple files - create zip
    try {
      setStatusMessage(`Preparing ${count} ${downloadType} image(s) for download...`);
      
      const zip = new JSZip();
      let completed = 0;
      
      // Fetch all images and add to zip
      for (const imageId of selectedImageIds) {
        const image = images.find(img => img.id === imageId);
        
        // Skip if requested type not available
        if (downloadType === 'staged' && !image?.staged_url) continue;
        if (downloadType === 'original' && !image) continue;
        
        try {
          // Get presigned URL
          const params = new URLSearchParams({ kind: downloadType });
          const res = await apiFetch<{ url: string }>(`/v1/images/${imageId}/presign?${params.toString()}`);
          
          if (res?.url) {
            // Fetch the image as blob
            const response = await fetch(res.url);
            const blob = await response.blob();
            
            // Generate filename
            const extension = blob.type.split('/')[1] || 'jpg';
            const roomType = image?.room_type || 'image';
            const filename = `${roomType.replace(/\s+/g, '-')}_${downloadType}_${imageId.substring(0, 8)}.${extension}`;
            
            // Add to zip
            zip.file(filename, blob);
            
            completed++;
            setStatusMessage(`Preparing ${completed}/${count} images...`);
          }
        } catch (err) {
          console.error(`Failed to fetch image ${imageId}:`, err);
        }
      }
      
      if (completed === 0) {
        setStatusMessage('No images available for download');
        setTimeout(() => setStatusMessage(''), 3000);
        return;
      }
      
      // Generate zip file
      setStatusMessage('Creating zip file...');
      const zipBlob = await zip.generateAsync({ type: 'blob' });
      
      // Trigger download
      const link = document.createElement('a');
      link.href = URL.createObjectURL(zipBlob);
      const projectName = selectedProject?.name.replace(/\s+/g, '-') || 'images';
      link.download = `${projectName}_${downloadType}_${new Date().toISOString().split('T')[0]}.zip`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      
      // Clean up
      URL.revokeObjectURL(link.href);
      
      setStatusMessage(`Downloaded ${completed} image(s) as zip file`);
      setTimeout(() => setStatusMessage(''), 3000);
      
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(`Download failed: ${message}`);
      setTimeout(() => setStatusMessage(''), 5000);
    }
  }

  async function loadProjects() {
    setLoadingProjects(true);
    setStatusMessage("Loading projects...");
    try {
      const res = await apiFetch<ProjectListResponse>("/v1/projects");
      const list = res.projects ?? [];
      setProjects(list);
      if (list.length === 0) {
        setSelectedProjectId("");
        setImages([]);
        setStatusMessage("No projects found. Create one from the Upload page.");
        return;
      }

      if (!selectedProjectId || !list.some((project) => project.id === selectedProjectId)) {
        setSelectedProjectId(list[0].id);
      }
      setStatusMessage("");
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(message);
    } finally {
      setLoadingProjects(false);
    }
  }

  async function loadImages(projectId: string, isBackground = false) {
    if (!projectId) {
      setImages([]);
      setImageUrls({});
      return;
    }
    
    // Only show loading UI for initial loads, not background refreshes
    if (!isBackground) {
      setLoadingImages(true);
      setStatusMessage("Loading images...");
    }
    
    try {
      const res = await apiFetch<ImageListResponse>(`/v1/projects/${projectId}/images`);
      const list = res.images ?? [];
      setImages(list);
      
      // Only clear selection on initial load
      if (!isBackground) {
        setSelectedImageIds(new Set());
        setStatusMessage(list.length === 0 ? "No images found for this project yet." : "");
      }
      
      // Prefetch image URLs for display
      if (list.length > 0) {
        prefetchImageUrls(list);
      }
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : String(err);
      if (!isBackground) {
        setStatusMessage(message);
        setImages([]);
        setImageUrls({});
      }
      // Silently fail for background refreshes
    } finally {
      if (!isBackground) {
        setLoadingImages(false);
      }
    }
  }

  useEffect(() => {
    loadProjects();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (!selectedProjectId) {
      setImages([]);
      return;
    }
    loadImages(selectedProjectId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId]);

  // Manual refresh function
  const handleManualRefresh = useCallback(async () => {
    if (!selectedProjectId || isRefreshing) return;
    
    setIsRefreshing(true);
    try {
      await loadImages(selectedProjectId, false);
    } finally {
      setIsRefreshing(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedProjectId, isRefreshing]);

  // Delete image function
  const deleteImage = useCallback(async (imageId: string) => {
    if (!confirm('Are you sure you want to delete this image? It will not be recoverable.')) {
      return;
    }

    try {
      setStatusMessage('Deleting image...');
      await apiFetch(`/v1/images/${imageId}`, {
        method: 'DELETE',
      });
      
      // Remove from local state
      setImages(prev => prev.filter(img => img.id !== imageId));
      setImageUrls(prev => {
        const newUrls = { ...prev };
        delete newUrls[imageId];
        return newUrls;
      });
      setSelectedImageIds(prev => {
        const newSet = new Set(prev);
        newSet.delete(imageId);
        return newSet;
      });
      if (focusedImageId === imageId) {
        setFocusedImageId(null);
      }
      
      setStatusMessage('Image deleted successfully');
      setTimeout(() => setStatusMessage(''), 3000);
    } catch (err: unknown) {
      console.error('Failed to delete image:', err);
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(`Delete failed: ${message}`);
    }
  }, [focusedImageId]);

  // Delete selected images
  const deleteSelected = useCallback(async () => {
    const count = selectedImageIds.size;
    if (count === 0) return;

    if (!confirm(`Are you sure you want to delete ${count} image${count > 1 ? 's' : ''}? They will not be recoverable.`)) {
      return;
    }

    try {
      setStatusMessage(`Deleting ${count} image${count > 1 ? 's' : ''}...`);
      
      // Delete all selected images
      await Promise.all(
        Array.from(selectedImageIds).map(imageId =>
          apiFetch(`/v1/images/${imageId}`, { method: 'DELETE' })
        )
      );
      
      // Remove from local state
      setImages(prev => prev.filter(img => !selectedImageIds.has(img.id)));
      setImageUrls(prev => {
        const newUrls = { ...prev };
        selectedImageIds.forEach(id => delete newUrls[id]);
        return newUrls;
      });
      setSelectedImageIds(new Set());
      setFocusedImageId(null);
      
      setStatusMessage(`Successfully deleted ${count} image${count > 1 ? 's' : ''}`);
      setTimeout(() => setStatusMessage(''), 3000);
    } catch (err: unknown) {
      console.error('Failed to delete images:', err);
      const message = err instanceof Error ? err.message : String(err);
      setStatusMessage(`Delete failed: ${message}`);
    }
  }, [selectedImageIds]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ignore if typing in an input field
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
        return;
      }

      if (images.length === 0) return;

      const currentIndex = focusedImageId 
        ? images.findIndex(img => img.id === focusedImageId)
        : -1;

      switch (e.key) {
        case 'ArrowRight':
        case 'ArrowDown': {
          e.preventDefault();
          const nextIndex = viewMode === 'grid'
            ? (currentIndex + 1) % images.length
            : (currentIndex + 1) % images.length;
          setFocusedImageId(images[nextIndex].id);
          break;
        }
        case 'ArrowLeft':
        case 'ArrowUp': {
          e.preventDefault();
          const prevIndex = currentIndex <= 0 
            ? images.length - 1 
            : currentIndex - 1;
          setFocusedImageId(images[prevIndex].id);
          break;
        }
        case ' ': { // Space bar
          e.preventDefault();
          if (focusedImageId) {
            toggleImageSelection(focusedImageId);
          } else if (images.length > 0) {
            setFocusedImageId(images[0].id);
          }
          break;
        }
        case 'Delete':
        case 'Backspace': {
          e.preventDefault();
          if (selectedImageIds.size > 0) {
            deleteSelected();
          } else if (focusedImageId) {
            deleteImage(focusedImageId);
          }
          break;
        }
        case 'Escape': {
          e.preventDefault();
          setFocusedImageId(null);
          setSelectedImageIds(new Set());
          break;
        }
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [images, focusedImageId, selectedImageIds, viewMode, toggleImageSelection, deleteImage, deleteSelected]);

  // Set initial focus when images load
  useEffect(() => {
    if (images.length > 0 && !focusedImageId) {
      setFocusedImageId(images[0].id);
    }
  }, [images, focusedImageId]);

  // Determine if we should be polling based on current images
  const shouldPoll = useMemo(() => {
    return selectedProjectId && images.some(
      img => img.status === 'queued' || img.status === 'processing'
    );
  }, [selectedProjectId, images]);

  // Auto-polling for processing images
  useEffect(() => {
    // Clear any existing interval
    if (pollingInterval) {
      clearInterval(pollingInterval);
      setPollingInterval(null);
    }

    // Only set up polling if needed
    if (!shouldPoll) {
      pollingStartTimeRef.current = null;
      return;
    }

    // Start polling timer if not already started
    if (!pollingStartTimeRef.current) {
      pollingStartTimeRef.current = Date.now();
    }
    const MAX_POLLING_DURATION = 5 * 60 * 1000; // 5 minutes

    // Set up polling interval that checks current image states
    const interval = setInterval(() => {
      // Check if max polling duration exceeded
      if (pollingStartTimeRef.current && 
          Date.now() - pollingStartTimeRef.current > MAX_POLLING_DURATION) {
        console.warn('Max polling duration exceeded, stopping auto-refresh');
        clearInterval(interval);
        setPollingInterval(null);
        pollingStartTimeRef.current = null;
        return;
      }

      const hasProcessingImages = imagesRef.current.some(
        img => img.status === 'queued' || img.status === 'processing'
      );

      if (hasProcessingImages) {
        loadImages(selectedProjectId, true);
      } else {
        // All images are done processing, stop polling
        clearInterval(interval);
        setPollingInterval(null);
        pollingStartTimeRef.current = null;
      }
    }, 3000);

    setPollingInterval(interval);

    // Cleanup on unmount or when shouldPoll changes
    return () => {
      if (interval) {
        clearInterval(interval);
      }
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [shouldPoll, selectedProjectId]);

  return (
    <div className="space-y-4 sm:space-y-6 lg:space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-2xl sm:text-3xl font-bold mb-2">
          <span className="gradient-text">Image Gallery</span>
        </h1>
        <p className="text-sm sm:text-base text-gray-600 dark:text-gray-400">
          View and manage your virtually staged images across all projects
        </p>
      </div>

      {/* Project Selection */}
      <div className="card">
        <div className="card-header">
          <div className="flex items-center gap-2">
            <FolderOpen className="h-5 w-5 text-blue-600" />
            <span>Project</span>
          </div>
        </div>
        <div className="card-body">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <select
                className="input"
                value={selectedProjectId}
                onChange={(e) => {
                  setSelectedProjectId(e.target.value);
                  setStatusMessage("");
                }}
                disabled={loadingProjects}
              >
                <option value="">Select a project...</option>
                {projects.map((project) => (
                  <option key={project.id} value={project.id}>
                    {project.name}
                  </option>
                ))}
              </select>
            </div>
            <button 
              type="button" 
              className="btn btn-ghost w-full sm:w-auto" 
              onClick={loadProjects}
              disabled={loadingProjects}
            >
              <RefreshCw className={cn("h-4 w-4", loadingProjects && "animate-spin")} />
              <span className="sm:inline">Refresh</span>
            </button>
          </div>
          {selectedProject && (
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 sm:gap-0 mt-3">
              <div className="flex items-center gap-2 sm:gap-3 flex-wrap">
                <p className="text-xs sm:text-sm text-gray-600 dark:text-gray-400">
                  Viewing <span className="font-medium">{selectedProject.name}</span> • {images.length} image{images.length !== 1 ? 's' : ''}
                </p>
                {pollingInterval && (
                  <span className="flex items-center gap-1.5 text-xs text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-950/30 px-2 py-1 rounded-full">
                    <Loader2 className="h-3 w-3 animate-spin" />
                    Auto-updating
                  </span>
                )}
              </div>
              <button
                onClick={handleManualRefresh}
                disabled={isRefreshing || loadingImages}
                className="btn btn-secondary text-xs sm:text-sm"
                title="Manually refresh images"
              >
                <RefreshCw className={cn("h-4 w-4", isRefreshing && "animate-spin")} />
                <span className="hidden sm:inline">Refresh</span>
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Toolbar */}
      {images.length > 0 && (
        <div className="card">
          <div className="card-body">
            <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 sm:gap-4">
              {/* Selection Controls */}
              <div className="flex items-center gap-2 sm:gap-3 flex-wrap">
                <button
                  onClick={selectAll}
                  className="btn btn-ghost text-xs sm:text-sm px-3 sm:px-4"
                  disabled={images.length === 0}
                >
                  <Check className="h-4 w-4" />
                  <span className="hidden xs:inline">Select All</span> ({images.length})
                </button>
                <button
                  onClick={clearSelection}
                  className="btn btn-ghost text-xs sm:text-sm px-3 sm:px-4"
                  disabled={selectedImageIds.size === 0}
                >
                  <X className="h-4 w-4" />
                  <span className="hidden xs:inline">Clear</span>
                </button>
                {selectedImageIds.size > 0 && (
                  <span className="text-xs sm:text-sm font-medium text-blue-600 dark:text-blue-400">
                    {selectedImageIds.size} selected
                  </span>
                )}
              </div>

              {/* View Mode & Actions */}
              <div className="flex items-center gap-2 flex-wrap">
                {selectedImageIds.size > 0 && (
                  <>
                    {/* Download Type Toggle */}
                    <div className="flex rounded-lg border border-gray-200 dark:border-gray-700 p-1 mr-2">
                      <button
                        onClick={() => setDownloadType('staged')}
                        className={cn(
                          "px-3 py-1 text-xs rounded transition-colors",
                          downloadType === 'staged'
                            ? "bg-blue-100 text-blue-600 font-medium"
                            : "text-gray-600 hover:bg-gray-100"
                        )}
                        title="Download staged (AI-enhanced) images"
                      >
                        Staged
                      </button>
                      <button
                        onClick={() => setDownloadType('original')}
                        className={cn(
                          "px-3 py-1 text-xs rounded transition-colors",
                          downloadType === 'original'
                            ? "bg-blue-100 text-blue-600 font-medium"
                            : "text-gray-600 hover:bg-gray-100"
                        )}
                        title="Download original (unprocessed) images"
                      >
                        Original
                      </button>
                    </div>
                    <button
                      onClick={downloadSelected}
                      className="btn btn-secondary text-xs sm:text-sm"
                    >
                      <Download className="h-4 w-4" />
                      <span className="hidden sm:inline">Download {downloadType}</span> ({selectedImageIds.size})
                    </button>
                    <button
                      onClick={deleteSelected}
                      className="btn btn-ghost text-red-600 hover:bg-red-50 dark:hover:bg-red-950/30 text-xs sm:text-sm"
                    >
                      <Trash2 className="h-4 w-4" />
                      <span className="hidden sm:inline">Delete</span> ({selectedImageIds.size})
                    </button>
                  </>
                )}
                <div className="flex rounded-lg border border-gray-200 dark:border-gray-700 p-1">
                  <button
                    onClick={() => setViewMode('grid')}
                    className={cn(
                      "p-2 rounded transition-colors touch-manipulation",
                      viewMode === 'grid' 
                        ? "bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400" 
                        : "text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800"
                    )}
                    title="Grid view"
                  >
                    <Grid3x3 className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => setViewMode('list')}
                    className={cn(
                      "p-2 rounded transition-colors touch-manipulation",
                      viewMode === 'list' 
                        ? "bg-blue-100 dark:bg-blue-900/50 text-blue-600 dark:text-blue-400" 
                        : "text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-800"
                    )}
                    title="List view"
                  >
                    <List className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Loading State */}
      {loadingImages && (
        <div className="flex flex-col items-center justify-center py-16">
          <Loader2 className="h-12 w-12 text-blue-600 animate-spin mb-4" />
          <p className="text-gray-600">Loading images...</p>
        </div>
      )}

      {/* Empty State */}
      {!loadingImages && images.length === 0 && selectedProjectId && (
        <div className="card">
          <div className="card-body">
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <div className="rounded-full bg-gray-100 p-4 mb-4">
                <ImageIcon className="h-12 w-12 text-gray-400" />
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">No images yet</h3>
              <p className="text-gray-600 mb-4 max-w-md">
                Upload your first property image to get started with AI-powered virtual staging
              </p>
              <a href="/upload" className="btn btn-primary">
                Upload Image
              </a>
            </div>
          </div>
        </div>
      )}

      {/* Grid View */}
      {!loadingImages && images.length > 0 && viewMode === 'grid' && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 sm:gap-6">
          {images.map((image, index) => {
            const urls = imageUrls[image.id];
            const stagedSrc = urls?.staged ?? null;
            const originalSrc = urls?.original ?? null;
            // Prioritize first 3 images for LCP optimization
            const isPriority = index < 3;

            return (
              <div
                key={image.id}
                ref={(el) => registerImageObserver(el, image.id)}
                className={cn(
                  "card group cursor-pointer transition-all duration-200",
                  selectedImageIds.has(image.id) && "ring-2 ring-blue-500 shadow-xl",
                  focusedImageId === image.id && !selectedImageIds.has(image.id) && "ring-2 ring-blue-300 shadow-lg"
                )}
                onClick={() => toggleImageSelection(image.id)}
                onMouseEnter={() => handleImageHover(image.id)}
                onMouseLeave={() => setHoveredImageId(null)}
              >
                <div className="relative aspect-video bg-gray-100 dark:bg-gray-800 overflow-hidden rounded-t-2xl">
                  {/* Processing Overlay - Show FIRST if processing */}
                  {(image.status === 'queued' || image.status === 'processing') ? (
                    <div className="absolute inset-0 bg-gradient-to-br from-gray-800 to-gray-900 flex flex-col items-center justify-center text-white z-20">
                      <Loader2 className="h-12 w-12 animate-spin mb-3" />
                      <p className="text-sm font-medium capitalize">{image.status}</p>
                      <p className="text-xs text-gray-300 mt-1">
                        {image.status === 'queued' ? 'Waiting to process...' : 'AI staging in progress...'}
                      </p>
                    </div>
                  ) : (
                    /* Image Preview - Show staged by default, original on hover */
                    urls ? (
                      <>
                      {/* Staged Image (shown by default) */}
                      {typeof stagedSrc === "string" && (
                        <NextImage
                          src={stagedSrc}
                          alt="Staged"
                          fill
                          unoptimized
                          className={cn(
                            "absolute inset-0 object-cover transition-all duration-300",
                            hoveredImageId === image.id ? "opacity-0" : "opacity-100 group-hover:scale-105"
                          )}
                          priority={isPriority}
                          loading={isPriority ? undefined : "lazy"}
                          sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                        />
                      )}
                      {/* Original Image (shown on hover) */}
                      {typeof originalSrc === "string" && (
                        <NextImage
                          src={originalSrc}
                          alt="Original"
                          fill
                          unoptimized
                          className={cn(
                            "absolute inset-0 object-cover transition-opacity duration-300",
                            hoveredImageId === image.id ? "opacity-100" : "opacity-0"
                          )}
                          loading="lazy"
                          sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
                        />
                      )}
                      {/* Fallback if no images loaded yet */}
                      {!stagedSrc && !originalSrc && (
                        <div className="flex items-center justify-center h-full">
                          <Loader2 className="h-16 w-16 text-gray-300 animate-spin" />
                        </div>
                      )}
                    </>
                  ) : (
                    <div className="flex items-center justify-center h-full">
                      <Loader2 className="h-16 w-16 text-gray-300 animate-spin" />
                    </div>
                    )
                  )}

                  {/* Selection Indicator */}
                  <div className={cn(
                  "absolute top-3 left-3 flex items-center justify-center h-6 w-6 rounded-full border-2 transition-all",
                  selectedImageIds.has(image.id)
                    ? "bg-blue-600 border-blue-600"
                    : "bg-white border-white group-hover:border-blue-400"
                )}>
                  {selectedImageIds.has(image.id) && (
                    <Check className="h-4 w-4 text-white" />
                  )}
                </div>

                {/* Action Buttons */}
                <div className="absolute top-3 right-3 flex items-center gap-2">
                  <span className={cn(
                    "badge",
                    `badge-status-${image.status}`
                  )}>
                    {image.status}
                  </span>
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      deleteImage(image.id);
                    }}
                    className="p-1.5 bg-red-600 hover:bg-red-700 text-white rounded-lg opacity-0 group-hover:opacity-100 sm:opacity-100 md:opacity-0 transition-opacity touch-manipulation"
                    title="Delete image"
                    aria-label="Delete image"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>

                {/* Original/Staged Indicator */}
                {imageUrls[image.id]?.original && imageUrls[image.id]?.staged && (
                  <div className="absolute bottom-3 left-3 z-10">
                    <span className="badge bg-black/70 text-white text-xs">
                      {hoveredImageId === image.id ? "Original" : "Staged"}
                    </span>
                  </div>
                )}

                {/* Quick Actions Bar - Bottom */}
                <div className="absolute bottom-0 inset-x-0 bg-gradient-to-t from-black/70 to-transparent opacity-0 group-hover:opacity-100 sm:opacity-100 md:opacity-0 transition-opacity p-2 sm:p-3 flex items-center justify-center gap-2">
                  <button
                    onClick={(e) => {
                      e.stopPropagation();
                      openPreview(image.id, 'original');
                    }}
                    className="btn btn-secondary btn-sm"
                    title="View Original"
                  >
                    <ExternalLink className="h-3 w-3" />
                    <span className="hidden sm:inline">Original</span>
                  </button>
                  {image.staged_url && (
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        openPreview(image.id, 'staged');
                      }}
                      className="btn btn-primary btn-sm"
                      title="View Staged"
                    >
                      <ExternalLink className="h-3 w-3" />
                      <span className="hidden sm:inline">Staged</span>
                    </button>
                  )}
                </div>
              </div>

              {/* Card Content */}
              <div className="p-3 sm:p-4 space-y-2">
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-gray-900 truncate">
                      {image.room_type || 'Untitled'}
                    </p>
                    <p className="text-xs text-gray-500">
                      {formatRelativeTime(image.updated_at)}
                    </p>
                  </div>
                  {image.style && (
                    <span className="text-xs px-2 py-1 bg-gray-100 text-gray-700 rounded-full">
                      {image.style}
                    </span>
                  )}
                </div>

                {image.error && (
                  <div className="flex items-start gap-2 text-xs text-red-600 bg-red-50 p-2 rounded">
                    <AlertCircle className="h-3 w-3 mt-0.5 flex-shrink-0" />
                    <span className="line-clamp-2">{image.error}</span>
                  </div>
                )}
              </div>
            </div>
          );
          })}
        </div>
      )}

      {/* List View */}
      {!loadingImages && images.length > 0 && viewMode === 'list' && (
        <div className="card">
          <div className="card-body p-0">
            <div className="overflow-x-auto -mx-4 sm:mx-0 scroll-smooth-mobile">
              <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
                <thead className="bg-gray-50 dark:bg-gray-800">
                  <tr>
                    <th className="w-12 px-4 py-3"></th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Preview</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Details</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Status</th>
                    <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Updated</th>
                    <th className="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Actions</th>
                  </tr>
                </thead>
                <tbody className="bg-white dark:bg-gray-900 divide-y divide-gray-200 dark:divide-gray-700">
                  {images.map((image) => {
                    const thumbSrc = imageUrls[image.id]?.staged ?? null;

                    return (
                      <tr
                        key={image.id}
                        className={cn(
                          "hover:bg-gray-50 dark:hover:bg-gray-800 transition-colors cursor-pointer",
                          selectedImageIds.has(image.id) && "bg-blue-50 dark:bg-blue-950/30",
                          focusedImageId === image.id && !selectedImageIds.has(image.id) && "bg-blue-50/50 dark:bg-blue-950/10"
                        )}
                        onClick={() => toggleImageSelection(image.id)}
                      >
                      <td className="px-4 py-4">
                        <div
                          className={cn(
                            "flex items-center justify-center h-5 w-5 rounded border-2 transition-all",
                            selectedImageIds.has(image.id)
                              ? "bg-blue-600 border-blue-600"
                              : "border-gray-300"
                          )}
                        >
                          {selectedImageIds.has(image.id) && (
                            <Check className="h-3 w-3 text-white" />
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <div className="relative h-16 w-24 rounded-lg overflow-hidden bg-gray-100 dark:bg-gray-800">
                          {/* Processing State - Single spinner */}
                          {(image.status === 'queued' || image.status === 'processing') ? (
                            <div className="flex items-center justify-center h-full bg-gradient-to-br from-gray-700 to-gray-800">
                              <Loader2 className="h-6 w-6 text-white animate-spin" />
                            </div>
                          ) : (
                            /* Normal image display */
                            typeof thumbSrc === "string" ? (
                              <NextImage
                                src={thumbSrc}
                                alt="Preview"
                                width={96}
                                height={64}
                                unoptimized
                                className="h-full w-full object-cover"
                              />
                            ) : imageUrls[image.id] ? (
                              <div className="flex items-center justify-center h-full">
                                <ImageIcon className="h-8 w-8 text-gray-300" />
                              </div>
                            ) : (
                              <div className="flex items-center justify-center h-full">
                                <Loader2 className="h-6 w-6 text-gray-300 animate-spin" />
                              </div>
                            )
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <div>
                          <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                            {image.room_type || 'Untitled'}
                          </p>
                          {image.style && (
                            <p className="text-xs text-gray-500 dark:text-gray-400">{image.style}</p>
                          )}
                        </div>
                      </td>
                      <td className="px-4 py-4">
                        <span className={cn("badge badge-status-" + image.status)}>
                          {image.status}
                        </span>
                        {image.error && (
                          <p className="text-xs text-red-600 dark:text-red-400 mt-1 max-w-xs truncate">
                            {image.error}
                          </p>
                        )}
                      </td>
                      <td className="px-4 py-4 text-sm text-gray-600 dark:text-gray-400">
                        {formatRelativeTime(image.updated_at)}
                      </td>
                      <td className="px-4 py-4 text-right">
                        <div className="flex items-center justify-end gap-2">
                          <button
                            onClick={() => {
                              openPreview(image.id, 'original');
                            }}
                            className="text-gray-600 dark:text-gray-400 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
                            title="View Original"
                          >
                            <ExternalLink className="h-4 w-4" />
                          </button>
                          {image.staged_url && (
                            <button
                              onClick={() => {
                                openPreview(image.id, 'staged');
                              }}
                              className="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
                              title="View Staged"
                            >
                              <CheckCircle2 className="h-4 w-4" />
                          </button>
                          )}
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              deleteImage(image.id);
                            }}
                            className="text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 transition-colors"
                            title="Delete image"
                          >
                            <Trash2 className="h-4 w-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                  })}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}

      {/* Status Message */}
      {statusMessage && (
        <div className="fixed bottom-4 right-4 left-4 sm:left-auto bg-white dark:bg-gray-900 shadow-lg rounded-lg p-3 sm:p-4 border border-gray-200 dark:border-gray-700 max-w-sm mx-auto sm:mx-0 animate-in z-50">
          <p className="text-xs sm:text-sm text-gray-700 dark:text-gray-300">{statusMessage}</p>
        </div>
      )}

      {/* Image Preview Lightbox */}
      {previewImageId && (() => {
        const currentImage = previewImages[currentPreviewIndex];
        if (!currentImage) return null;

        const urls = imageUrls[currentImage.id];
        const displayUrl = previewMode === 'staged' ? urls?.staged : urls?.original;
        const canShowStaged = currentImage.staged_url && urls?.staged;

        return (
          <div 
            className="fixed inset-0 z-50 bg-black/95 flex items-center justify-center touch-manipulation"
            onClick={closePreview}
          >
            {/* Close Button */}
            <button
              onClick={closePreview}
              className="absolute top-2 right-2 sm:top-4 sm:right-4 z-10 p-2 sm:p-2.5 bg-white/10 hover:bg-white/20 active:bg-white/30 rounded-full transition-colors touch-manipulation"
              title="Close (Esc)"
              aria-label="Close preview"
            >
              <XCircle className="h-6 w-6 sm:h-8 sm:w-8 text-white" />
            </button>

            {/* Image Info Overlay - Top */}
            <div className="absolute top-2 left-2 sm:top-4 sm:left-4 z-10 bg-black/60 backdrop-blur-sm px-3 py-1.5 sm:px-4 sm:py-2 rounded-lg">
              <p className="text-white text-xs sm:text-sm font-medium">
                {currentPreviewIndex + 1} / {previewImages.length}
              </p>
              {selectedImageIds.size > 0 && (
                <p className="text-white/70 text-xs mt-0.5 sm:mt-1 hidden sm:block">
                  Showing selected images only
                </p>
              )}
            </div>

            {/* Navigation - Previous */}
            {currentPreviewIndex > 0 && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  goToPreviousImage();
                }}
                className="absolute left-2 sm:left-4 top-1/2 -translate-y-1/2 z-10 p-2 sm:p-3 bg-white/10 hover:bg-white/20 active:bg-white/30 rounded-full transition-colors touch-manipulation"
                title="Previous (←)"
                aria-label="Previous image"
              >
                <ChevronLeft className="h-6 w-6 sm:h-8 sm:w-8 text-white" />
              </button>
            )}

            {/* Navigation - Next */}
            {currentPreviewIndex < previewImages.length - 1 && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  goToNextImage();
                }}
                className="absolute right-2 sm:right-4 top-1/2 -translate-y-1/2 z-10 p-2 sm:p-3 bg-white/10 hover:bg-white/20 active:bg-white/30 rounded-full transition-colors touch-manipulation"
                title="Next (→)"
                aria-label="Next image"
              >
                <ChevronRight className="h-6 w-6 sm:h-8 sm:w-8 text-white" />
              </button>
            )}

            {/* Main Image Container */}
            <div 
              className="relative max-w-7xl max-h-[90vh] w-full h-full flex items-center justify-center px-12 sm:px-16 md:px-20"
              onClick={(e) => e.stopPropagation()}
            >
              {displayUrl ? (
                <div className="relative w-full h-full flex items-center justify-center">
                  <NextImage
                    src={displayUrl}
                    alt={previewMode === 'staged' ? 'Staged' : 'Original'}
                    fill
                    className="object-contain"
                    unoptimized
                  />
                  
                  {/* Image Type Overlay */}
                  <div className="absolute top-2 right-2 sm:top-4 sm:right-4 bg-black/70 backdrop-blur-sm px-3 py-1.5 sm:px-4 sm:py-2 rounded-lg">
                    <p className="text-white text-xs sm:text-sm font-semibold uppercase tracking-wide">
                      {previewMode === 'staged' ? 'Staged' : 'Original'}
                    </p>
                    {previewMode === 'staged' && currentImage.style && (
                      <p className="text-white/80 text-xs mt-0.5 sm:mt-1 hidden sm:block">
                        {currentImage.style}
                      </p>
                    )}
                  </div>
                </div>
              ) : (
                <div className="flex flex-col items-center gap-4 text-white">
                  <Loader2 className="h-12 w-12 animate-spin" />
                  <p>Loading image...</p>
                </div>
              )}
            </div>

            {/* Bottom Controls */}
            <div className="absolute bottom-4 sm:bottom-8 left-4 right-4 sm:left-1/2 sm:right-auto sm:-translate-x-1/2 z-10 flex flex-col sm:flex-row items-stretch sm:items-center gap-2 sm:gap-3">
              {/* Toggle Original/Staged */}
              {canShowStaged && (
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    togglePreviewMode();
                  }}
                  className="px-4 py-2.5 sm:px-6 sm:py-3 bg-white/10 hover:bg-white/20 active:bg-white/30 backdrop-blur-sm rounded-lg transition-colors flex items-center justify-center gap-2 touch-manipulation"
                  title="Toggle View (Space or T)"
                >
                  <span className="text-white text-sm sm:text-base font-medium">
                    {previewMode === 'original' ? 'Show Staged' : 'Show Original'}
                  </span>
                </button>
              )}

              {/* Image Details */}
              <div className="px-3 py-2 sm:px-4 sm:py-3 bg-black/60 backdrop-blur-sm rounded-lg">
                <p className="text-white text-xs sm:text-sm">
                  {currentImage.room_type || 'Untitled'}
                </p>
                <p className="text-white/60 text-xs mt-0.5 sm:mt-1">
                  {formatRelativeTime(currentImage.updated_at)}
                </p>
              </div>
            </div>
          </div>
        );
      })()}
    </div>
  );
}
